package tools

import (
	"context"
	"fmt"

	"github.com/obot-platform/nanobot/pkg/expr"
	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
	"github.com/obot-platform/nanobot/pkg/uuid"
)

type flowContext struct {
	ctx     context.Context
	opt     mcp.CallOption
	env     map[string]string
	data    map[string]any
	outputs map[string]any
}

func (r *Service) startFlow(ctx context.Context, flowName string, args any, opt CallOptions) (*mcp.CallToolResult, error) {
	flow, ok := r.config.Flows[flowName]
	if !ok {
		return nil, fmt.Errorf("failed to find flow %s in config", flowName)
	}

	data := map[string]any{
		"id":    uuid.String(),
		"flow":  flowName,
		"input": args,
	}

	fCtx := flowContext{
		ctx: ctx,
		opt: mcp.CallOption{
			ProgressToken: opt.ProgressToken,
		},
		env:     mcp.SessionFromContext(ctx).EnvMap(),
		data:    data,
		outputs: data,
	}

	return r.runSteps(fCtx, flow.Steps)
}

func (r *Service) runSteps(ctx flowContext, steps []types.Step) (*mcp.CallToolResult, error) {
	for i, step := range steps {
		if step.ID == "" {
			step.ID = uuid.String()
		}

		out, err := r.runStep(ctx, step)
		if err != nil {
			return nil, fmt.Errorf("failed to run step %d (%s): %w", i, step.ID, err)
		}

		if i == len(steps)-1 {
			// If this is the last step, we return the output directly.
			return out, nil
		}
	}

	return &mcp.CallToolResult{
		Content: make([]mcp.Content, 0),
	}, nil
}

func (r *Service) runStepEach(ctx flowContext, step types.Step, forEachData []any) (ret *mcp.CallToolResult, err error) {
	var (
		results     = make([]map[string]any, 0)
		itemVarName = "item"
	)

	if step.ForEachVar != "" {
		itemVarName = step.ForEachVar
	}

	oldVar, hadOldVar := ctx.data[itemVarName]

	for _, item := range forEachData {
		step.ForEach = nil
		ctx.data[itemVarName] = item // Set the current item in the context data.
		result, err := r.runStep(ctx, step)
		if err != nil {
			return nil, fmt.Errorf("failed to run forEach step %s: %w", step.ID, err)
		}
		results = append(results, toOutput(result))
	}

	if hadOldVar {
		ctx.data[itemVarName] = oldVar
	} else {
		delete(ctx.data, itemVarName)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			{
				StructuredContent: results,
			},
		},
	}, nil
}

func (r *Service) logFlowState(ctx flowContext) {
	if ctx.opt.ProgressToken == nil {
		return
	}
	session := mcp.SessionFromContext(ctx.ctx)
	if session == nil {
		return
	}

	_ = session.SendPayload(ctx.ctx, "notifications/progress", mcp.NotificationProgressRequest{
		ProgressToken: ctx.opt.ProgressToken,
		Data: map[string]any{
			"type": "nanobot/flow/state",
			"flow": ctx.data,
		},
	})
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

func toOutput(ret *mcp.CallToolResult) map[string]any {
	if ret == nil {
		return nil
	}

	output := map[string]any{
		"content": ret.Content,
		"isError": ret.IsError,
	}
	if ret.Content == nil {
		output["content"] = make([]mcp.Content, 0)
	}
	for i := len(ret.Content) - 1; i >= 0; i-- {
		if ret.Content[i].Text != "" {
			output["output"] = ret.Content[i].Text
		}
		if ret.Content[i].StructuredContent != nil {
			output["output"] = ret.Content[i].StructuredContent
		}
	}
	return output
}

func (r *Service) runStep(ctx flowContext, step types.Step) (ret *mcp.CallToolResult, err error) {
	defer func() {
		ctx.outputs[step.ID] = toOutput(ret)
		ctx.outputs["previous"] = ctx.outputs[step.ID]
		r.logFlowState(ctx)
	}()

	if step.ID == "" {
		step.ID = uuid.String()
	}

	call := getCall(step)
	if call != "" && len(step.Steps) > 0 {
		return nil, fmt.Errorf("step %s cannot have both agent/tool/flow (%s) and steps defined (count: %d)",
			step.ID, call, len(step.Steps))
	}

	if call == "" && len(step.Steps) == 0 {
		return nil, fmt.Errorf("step %s must have either agent/tool/flow or steps defined. missing both", step.ID)
	}

	if step.If != "" {
		isTrue, err := expr.EvalBool(ctx.ctx, ctx.env, ctx.data, step.ForEach)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate if condition for step %s: %w", step.ID, err)
		}
		if !isTrue {
			return &mcp.CallToolResult{Content: make([]mcp.Content, 0)}, nil
		}
	}

	if step.ForEach != nil {
		forEachData, err := expr.EvalList(ctx.ctx, ctx.env, ctx.data, step.ForEach)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate forEach for step %s: %w", step.ID, err)
		}

		return r.runStepEach(ctx, step, forEachData)
	}

	inputData, err := expr.EvalObject(ctx.ctx, ctx.env, ctx.data, step.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate input for step %s: %w", step.ID, err)
	}

	if call != "" {
		ref := types.ParseToolRef(call)
		return r.Call(ctx.ctx, ref.Server, ref.Tool, inputData, CallOptions{
			ProgressToken: ctx.opt.ProgressToken,
			AgentOverride: step.Agent,
		})
	}

	return r.runSteps(ctx, step.Steps)
}
