package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"maps"
	"strings"
	"sync"

	"github.com/nanobot-ai/nanobot/pkg/complete"
	"github.com/nanobot-ai/nanobot/pkg/expr"
	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
	"github.com/nanobot-ai/nanobot/pkg/uuid"
	"golang.org/x/sync/errgroup"
)

type flowContext struct {
	ctx  context.Context
	opt  mcp.CallOption
	env  map[string]string
	data map[string]any
}

func (s *Service) getPrompt(ctx context.Context, prompt string, args map[string]string) (string, error) {
	server, prompt, ok := strings.Cut(prompt, "/")
	if !ok {
		prompt = server
	}

	promptResult, err := s.GetPrompt(ctx, server, prompt, args)
	if err != nil {
		return "", fmt.Errorf("failed to get prompt %s from server %s: %w", prompt, server, err)
	}

	for _, msg := range promptResult.Messages {
		if msg.Content.Text != "" {
			return msg.Content.Text, nil
		}
	}
	return "", nil
}

func (s *Service) newGlobals(ctx context.Context, vars map[string]any, opt ...CallOptions) map[string]any {
	session := mcp.SessionFromContext(ctx)
	attr := session.Attributes()
	data := map[string]any{}
	data["prompt"] = func(target string, args map[string]string) (string, error) {
		return s.getPrompt(ctx, target, args)
	}
	data["nanobot"] = attr
	data["call"] = func(target string, args map[string]any) (map[string]any, error) {
		return s.callFromScript(ctx, target, args, CallOptions{
			ProgressToken: complete.Complete(opt...).ProgressToken,
		})
	}
	servers := map[string]any{}
	data["servers"] = servers

	for k, v := range attr {
		cf, ok := v.(*clientFactory)
		if !ok {
			continue
		}
		serverName, ok := strings.CutPrefix(k, "clients/")
		if !ok {
			continue
		}
		var instructions string
		if cf.client != nil && cf.client.Session != nil {
			instructions = cf.client.Session.InitializeResult.Instructions
		}
		servers[serverName] = map[string]any{
			"instructions": instructions,
		}
	}
	maps.Copy(data, vars)
	return data
}

func (s *Service) callFromScript(ctx context.Context, target string, args any, opt CallOptions) (map[string]any, error) {
	server, tool, _ := strings.Cut(target, "/")
	ret, err := s.Call(ctx, server, tool, args, opt)
	if err != nil {
		return nil, err
	}
	return toOutput(ret), nil
}

func (s *Service) startFlow(ctx context.Context, config types.Config, flowName string, args any, opt CallOptions) (*types.CallResult, error) {
	flow, ok := config.Flows[flowName]
	if !ok {
		return nil, fmt.Errorf("failed to find flow %s in config", flowName)
	}

	data := s.newGlobals(ctx, map[string]any{
		"id":     uuid.String(),
		"flow":   flowName,
		"target": opt.Target,
	}, opt)

	if opt.ReturnOutput {
		data["output"] = args
	} else {
		data["input"] = args
	}

	fCtx := flowContext{
		ctx: ctx,
		opt: mcp.CallOption{
			ProgressToken: opt.ProgressToken,
		},
		env:  mcp.SessionFromContext(ctx).GetEnvMap(),
		data: data,
	}

	ret, err := s.runSteps(fCtx, flow.Steps)
	if returnErr := (*ErrReturn)(nil); errors.As(err, &returnErr) {
		ret = returnErr.Result
	} else if err != nil {
		return nil, err
	}

	if !ret.IsError {
		if opt.ReturnInput {
			return objectToResult(data["input"])
		} else if opt.ReturnOutput {
			return objectToResult(data["output"])
		}
	}

	return ret, nil
}

func (s *Service) runSteps(ctx flowContext, steps []types.Step) (*types.CallResult, error) {
	last := &types.CallResult{
		Content: make([]mcp.Content, 0),
	}
	for i, step := range steps {
		if step.ID == "" {
			step.ID = uuid.String()
		}

		out, err := s.runStep(ctx, step)
		if err != nil {
			return nil, fmt.Errorf("failed to run step %d (%s): %w", i, step.ID, err)
		}

		if out == nil {
			out = last
		} else {
			last = out
		}

		if last.IsError {
			break
		}

		if i == len(steps)-1 {
			// If this is the last step, we return the output directly.
			return out, nil
		}
	}

	return last, nil
}

func (s *Service) runStepForEach(ctx flowContext, step types.Step) (ret *types.CallResult, err error) {
	forEachData, err := expr.EvalList(ctx.ctx, ctx.env, ctx.data, step.ForEach)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate forEach for step %s: %w", step.ID, err)
	}

	return s.runStepEach(ctx, step, func(yield func(any) bool) {
		for _, val := range forEachData {
			if !yield(val) {
				return
			}
		}
	})
}

func (s *Service) runStepWhile(ctx flowContext, step types.Step) (ret *types.CallResult, err error) {
	var loopErr error
	defer func() {
		if err == nil && loopErr != nil {
			err = loopErr
		}
	}()

	step.Parallel = false
	return s.runStepEach(ctx, step, func(yield func(any) bool) {
		i := 0
		for {
			isTrue, err := expr.EvalBool(ctx.ctx, ctx.env, ctx.data, step.While)
			if err != nil {
				loopErr = err
				return
			}
			if !isTrue {
				return
			}
			if !yield(i) {
				return
			}
			i++
		}
	})
}

func (s *Service) runStepEach(ctx flowContext, step types.Step, forEachData iter.Seq[any]) (ret *types.CallResult, err error) {
	var (
		results     = make([]map[string]any, 0)
		itemVarName = "item"
		resultLock  sync.Mutex
		eg          errgroup.Group
	)

	if step.Parallel {
		eg.SetLimit(s.concurrency)
	} else {
		eg.SetLimit(1)
	}

	if step.ForEachVar != "" {
		itemVarName = step.ForEachVar
	}

	oldVar, hadOldVar := ctx.data[itemVarName]
	step.ForEach = nil
	step.While = ""

	for item := range forEachData {
		newCtx := ctx
		if step.Parallel {
			newCtx.data = maps.Clone(ctx.data)
		}
		ctx.data[itemVarName] = item
		eg.Go(func() error {
			result, err := s.runStep(newCtx, step)
			if err != nil {
				return fmt.Errorf("failed to run forEach step %s: %w", step.ID, err)
			}

			resultLock.Lock()
			defer resultLock.Unlock()
			results = append(results, toOutput(result))
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if hadOldVar {
		ctx.data[itemVarName] = oldVar
	} else {
		delete(ctx.data, itemVarName)
	}

	return &types.CallResult{
		StructuredContent: results,
	}, nil
}

func getCall(step types.Step) string {
	if step.Agent.Name != "" {
		return step.Agent.Name
	}
	if step.Tool != "" {
		return step.Tool
	}
	if step.Flow != "" {
		return step.Flow
	}
	return ""
}

func toOutput(ret *types.CallResult) map[string]any {
	if ret == nil {
		return nil
	}

	output := map[string]any{
		"content":           ret.Content,
		"isError":           ret.IsError,
		"structuredContent": ret.StructuredContent,
	}
	if ret.Content == nil {
		output["content"] = make([]mcp.Content, 0)
	}
	for i := len(ret.Content) - 1; i >= 0; i-- {
		if ret.Content[i].Text != "" {
			output["output"] = ret.Content[i].Text
		}
		if ret.StructuredContent != nil {
			output["output"] = ret.StructuredContent
		}
	}
	return output
}

type ErrReturn struct {
	Result *types.CallResult
}

func (e *ErrReturn) Error() string {
	d, _ := json.Marshal(e.Result)
	return fmt.Sprintf("return: %s", string(d))
}

func (s *Service) elicit(ctx flowContext, elicit types.Elicit) (*types.CallResult, error) {
	var (
		session  = mcp.SessionFromContext(ctx.ctx)
		request  mcp.ElicitRequest
		response mcp.ElicitResult
	)
	msg, err := expr.EvalString(ctx.ctx, ctx.env, ctx.data, elicit.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate elicit message: %w", err)
	}

	request.Message = msg

	if elicit.Input != nil {
		if err := json.Unmarshal(elicit.Input.ToSchema(), &request.RequestedSchema); err != nil {
			return nil, fmt.Errorf("failed to unmarshal elicit input schema: %w", err)
		}
	}

	if err := session.Exchange(ctx.ctx, "elicitation/create", &request, &response); err != nil {
		return nil, fmt.Errorf("failed to create elicit request: %w", err)
	}

	switch response.Action {
	case "accept":
		return objectToResult(response.Content)
	case "reject":
		obj, err := expr.EvalAny(ctx.ctx, ctx.env, ctx.data,
			complete.First(elicit.RejectResult, elicit.CancelResult,
				any(map[string]any{
					"action":  "reject",
					"message": "The user has declined the elicitation: <message>" + elicit.Message + "</message>",
				})))
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate reject result: %w", err)
		}
		v, err := objectToResult(obj)
		if err != nil {
			return nil, err
		}
		v.IsError = true
		return v, nil
	case "cancel":
		obj, err := expr.EvalAny(ctx.ctx, ctx.env, ctx.data,
			complete.First(elicit.CancelResult, elicit.RejectResult,
				any(map[string]any{
					"action":  "cancel",
					"message": "The user has cancelled the elicitation: <message>" + elicit.Message + "</message>",
				})))
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate cancel result: %w", err)
		}
		v, err := objectToResult(obj)
		if err != nil {
			return nil, err
		}
		v.IsError = true
		return v, nil
	default:
		return nil, fmt.Errorf("unknown elicit action: %s", response.Action)
	}
}

func objectToResult(val any) (*types.CallResult, error) {
	if str, ok := val.(string); ok {
		return &types.CallResult{
			Content: []mcp.Content{
				{
					Text: str,
				},
			},
		}, nil
	}
	text, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result value: %w", err)
	}
	return &types.CallResult{
		StructuredContent: val,
		Content: []mcp.Content{
			{
				Text: string(text),
			},
		},
	}, nil
}

func (s *Service) runStep(ctx flowContext, step types.Step) (ret *types.CallResult, err error) {
	defer func() {
		ctx.data[step.ID] = toOutput(ret)
		ctx.data["previous"] = ctx.data[step.ID]
	}()

	if step.ID == "" {
		step.ID = uuid.String()
	}

	call := getCall(step)
	if call != "" && len(step.Steps) > 0 {
		return nil, fmt.Errorf("step %s cannot have both agent/tool/flow (%s) and steps defined (count: %d)",
			step.ID, call, len(step.Steps))
	}

	for k, v := range step.Set {
		if v == nil {
			delete(ctx.data, k)
		} else {
			val, err := expr.EvalAny(ctx.ctx, ctx.env, ctx.data, v)
			if err != nil {
				return nil, err
			}
			ctx.data[k] = val
		}
	}

	if step.ForEach != nil {
		return s.runStepForEach(ctx, step)
	}

	if step.While != "" {
		return s.runStepWhile(ctx, step)
	}

	if step.If != "" {
		isTrue, err := expr.EvalBool(ctx.ctx, ctx.env, ctx.data, step.ForEach)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate if condition for step %s: %w", step.ID, err)
		}
		if !isTrue {
			if len(step.Else) == 0 {
				return nil, nil
			}
			return s.runSteps(ctx, step.Else)
		}
	}

	if step.Elicit != nil {
		return s.elicit(ctx, *step.Elicit)
	}

	if step.Return != nil {
		val, err := expr.EvalObject(ctx.ctx, ctx.env, ctx.data, step.Return)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate return for step %s: %w", step.ID, err)
		}
		result, err := objectToResult(val)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal return value for step %s: %w", step.ID, err)
		}
		return nil, &ErrReturn{
			Result: result,
		}
	}

	if step.Evaluate != nil {
		evalStr, ok := step.Evaluate.(string)
		if ok {
			step.Evaluate = fmt.Sprintf("(function(){\n%s\n)()", evalStr)
		}

		val, err := expr.EvalAny(ctx.ctx, ctx.env, ctx.data, step.Evaluate)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate expression for step %s: %w", step.ID, err)
		}
		return objectToResult(val)
	}

	inputData, err := expr.EvalObject(ctx.ctx, ctx.env, ctx.data, step.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate input for step %s: %w", step.ID, err)
	}

	if call != "" {
		ref := types.ParseToolRef(call)
		return s.Call(ctx.ctx, ref.Server, ref.Tool, inputData, CallOptions{
			ProgressToken: ctx.opt.ProgressToken,
			AgentOverride: step.Agent,
		})
	}

	return s.runSteps(ctx, step.Steps)
}
