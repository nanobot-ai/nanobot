import type { Content, ToolCall } from "~/lib/nanobot";
import { UIResourceRenderer } from "@mcp-ui/client";
import { SparklesIcon } from "./icons";
import { ChevronDownIcon, Hammer } from "lucide-react";
import { useState } from "react";
import { Markdown } from "./markdown";
import { cn, sanitizeText } from "~/lib/utils";
import { clsx } from "clsx";

function isValidJson(str: string) {
  try {
    JSON.parse(str);
    return true;
  } catch (e) {
    return false;
  }
}

function JsonTable({ jsonString }: { jsonString: string }) {
  const data = JSON.parse(jsonString);
  return (
    <table className="w-full">
      <tbody>
        {Object.entries(data).map(([key, value]) => (
          <tr key={key} className="border-b border-border/50 last:border-0">
            <td className="py-1 pr-4 text-muted-foreground">{key}</td>
            <td className="py-1">
              {typeof value === "string" ? value : JSON.stringify(value)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function ResourceTable({ resource }: { resource: Content["resource"] }) {
  return (
    <table className="w-full">
      <thead>
        <tr>
          <th className="text-left py-2 font-medium" colSpan={2}>
            Resource
          </th>
        </tr>
      </thead>
      <tbody>
        {Object.entries(resource ?? {}).map(([key, value]) => (
          <tr key={key} className="border-b border-border/50 last:border-0">
            <td className="py-1 pr-4 text-muted-foreground">{key}</td>
            <td className="py-1">
              {typeof value === "string" ? value : JSON.stringify(value)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function ToolCall({ toolCall }: { toolCall: ToolCall }) {
  const { name, arguments: args, output, target, targetType } = toolCall;
  const [isExpanded, setIsExpanded] = useState(false);

  const isCompleted = !!output;

  return (
    <>
      <div
        className={cn({
          "p-4 my-2": true,
          "border rounded-lg bg-muted/30": isExpanded,
        })}
      >
        <div className="flex flex-col gap-3">
          {/* Tool Call Header */}
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="flex items-center gap-2 text-sm font-medium w-full"
          >
            <div className="size-5 flex items-center rounded-full justify-center ring-1 shrink-0 ring-border bg-background">
              <Hammer size={12} />
            </div>
            <span
              className={cn(
                "font-semibold",
                isCompleted ? "text-foreground" : "text-muted-foreground",
              )}
            >
              {isCompleted ? "" : "Calling:"} {name}
            </span>
            {target && targetType && (
              <span className="text-xs text-muted-foreground">
                ({targetType}: {target})
              </span>
            )}
            <ChevronDownIcon
              size={16}
              className={cn(
                "ml-auto transition-transform",
                isExpanded ? "rotate-0" : "-rotate-90",
              )}
            />
          </button>

          {isExpanded && (
            <div className="animate-in fade-in-0 slide-in-from-top-2">
              {/* Arguments */}
              {args && (
                <div className="pl-7">
                  <div className="text-xs text-muted-foreground mb-1">
                    Arguments:
                  </div>

                  <div className="bg-background rounded p-2 text-sm font-mono overflow-x-auto">
                    {isValidJson(args) ? <JsonTable jsonString={args} /> : args}
                  </div>
                </div>
              )}

              {/* Output */}
              {output && (
                <div className="pl-7 mt-1">
                  <div className="text-xs text-muted-foreground mb-1">
                    Result{output.isError ? " (Error)" : ""}
                    {output.agent ? ` from ${output.agent}` : ""}
                    {output.model ? ` using {output.model}` : ""}:
                  </div>
                  <div
                    className={cn(
                      "rounded p-3 text-sm",
                      output.isError
                        ? "bg-destructive/10 border border-destructive/20"
                        : "bg-background",
                    )}
                  >
                    {output.content?.map((content, index) => {
                      if (content.type === "text" && content.text) {
                        return (
                          <div key={index} className="mb-2">
                            <div className="text-xs text-muted-foreground mb-1">
                              Text Output:
                            </div>
                            <Markdown>{sanitizeText(content.text)}</Markdown>
                          </div>
                        );
                      } else if (
                        content.type === "resource" &&
                        content.resource
                      ) {
                        return content.resource.uri?.startsWith("ui://") ? (
                          <div key={index} className="mb-2">
                            <div className="text-xs text-muted-foreground mb-1">
                              MCP-UI:
                            </div>
                            <UIResourceRenderer resource={content.resource} />
                          </div>
                        ) : (
                          <div key={index} className="mb-2">
                            <ResourceTable resource={content.resource} />
                          </div>
                        );
                      } else if (content.structuredContent) {
                        return (
                          <div key={index} className="mb-2">
                            <div className="text-xs text-muted-foreground mb-1">
                              Structured Output:
                            </div>
                            <pre className="whitespace-pre-wrap overflow-x-auto">
                              {JSON.stringify(
                                content.structuredContent,
                                null,
                                2,
                              )}
                            </pre>
                          </div>
                        );
                      }
                      return null;
                    })}
                  </div>
                </div>
              )}

              {/* Loading state for pending tool calls */}
              {!isCompleted && (
                <div className="pl-7 flex items-center gap-2 text-muted-foreground">
                  <div className="animate-pulse h-2 w-2 rounded-full bg-muted-foreground"></div>
                  <div className="animate-pulse h-2 w-2 rounded-full bg-muted-foreground animation-delay-200"></div>
                  <div className="animate-pulse h-2 w-2 rounded-full bg-muted-foreground animation-delay-400"></div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
      {!isExpanded &&
        (
          output?.content?.flatMap((content) => {
            if (
              content.type === "resource" &&
              content.resource &&
              content.resource.uri?.startsWith("ui://") &&
              !content.resource.uri.startsWith("ui://widget/")
            ) {
              return [content.resource];
            }
            return [];
          }) || []
        ).map((resource, index) => (
          <UIResourceRenderer key={index} resource={resource} />
        ))}
    </>
  );
}
