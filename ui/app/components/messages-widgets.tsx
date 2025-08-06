import { useState } from "react";
import { Maximize, Minimize } from "lucide-react";
import type { Content } from "~/lib/nanobot";
import { UIResourceRenderer } from "@mcp-ui/client";

export function Widgets({ widgets }: { widgets: Content[] }) {
  const [fullScreenWidget, setFullScreenWidget] = useState<number | null>(null);

  const toggleFullScreen = (index: number) => {
    if (fullScreenWidget === index) {
      setFullScreenWidget(null);
    } else {
      setFullScreenWidget(index);
    }
  };

  if (fullScreenWidget !== null || widgets.length === 1) {
    const widget = widgets[widgets.length === 1 ? 0 : (fullScreenWidget ?? 0)];
    return (
      <div className="border border-[#e0e0e0] rounded-md p-6 flex flex-col h-full">
        <div className="flex justify-between items-center mb-4 pb-2 border-b">
          <h3 className="font-medium">{widget.resource?.uri}</h3>
          {widgets.length !== 1 && (
            <button
              className="p-1.5 hover:bg-[#f0f0f0] rounded-md cursor-pointer"
              onClick={() => setFullScreenWidget(null)}
              aria-label="Exit full screen"
            >
              <Minimize size={18} />
            </button>
          )}
        </div>
        <div className="flex-1">
          {widget.resource && (
            <UIResourceRenderer
              htmlProps={{
                iframeProps: {
                  className: "h-full",
                },
              }}
              resource={widget.resource}
            />
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-4">
      {widgets.map((widget, index) => (
        <div
          key={index}
          className="border border-[#e0e0e0] rounded-md p-4 flex flex-col"
        >
          <div className="flex justify-between items-center mb-4 pb-2 border-b">
            <h3 className="font-medium text-muted-foreground">
              {widget.resource?.name || widget.resource?.uri}
            </h3>
            <button
              className="p-1.5 hover:bg-[#f0f0f0] rounded-md cursor-pointer"
              onClick={() => toggleFullScreen(index)}
              aria-label="Full screen"
            >
              <Maximize size={18} />
            </button>
          </div>
          <div className="flex-1 h-full">
            {widget.resource && (
              <UIResourceRenderer
                htmlProps={{
                  iframeProps: {
                    className: "h-full",
                  },
                }}
                resource={widget.resource}
              />
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
