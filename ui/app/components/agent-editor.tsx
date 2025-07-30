import { useState, useMemo, useEffect, useRef } from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Textarea } from "./ui/textarea";
import { Switch } from "./ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "./ui/tabs";

import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Form, useSubmit, Link } from "react-router";
import type { ChatData, CustomAgent } from "~/lib/nanobot";
import { MinusIcon } from "lucide-react";

export function AgentEditor({
  customAgent,
  chat,
}: {
  customAgent: CustomAgent;
  chat: ChatData;
}) {
  const [isPublic, setIsPublic] = useState(customAgent.isPublic);
  const [serverUrls, setServerUrls] = useState<string[]>(
    customAgent?.mcpServers?.map((x) => x.url) || [""],
  );
  const [selectedModel, setSelectedModel] = useState(
    customAgent?.baseAgent || chat.currentAgent,
  );
  const formRef = useRef<HTMLFormElement>(null);
  const submit = useSubmit();
  const autoSaveTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Function to handle form changes
  const handleFormChange = () => {
    if (formRef.current) {
      const newFormData = new FormData(formRef.current);

      // Clear any existing timeout
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }

      // Set a new timeout for auto-save
      autoSaveTimeoutRef.current = setTimeout(() => {
        console.log("Auto-saving form...");
        submit(newFormData, {
          action: `/chat/${chat.id}/agent/${customAgent.id}/edit`,
          method: "put",
        });
      }, 1000); // 1 second delay
    }
  };

  // Clean up timeout on unmount
  useEffect(() => {
    return () => {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
    };
  }, []);

  const handleUrlChange = (index: number, value: string) => {
    const newUrls = [...serverUrls];
    newUrls[index] = value;
    setServerUrls(newUrls);

    // Trigger form change after state update
    setTimeout(handleFormChange, 0);
  };

  const addNewUrlInput = useMemo(() => {
    const lastUrl = serverUrls[serverUrls.length - 1];
    return lastUrl !== "";
  }, [serverUrls]);

  return (
    <Card>
      <Form
        action={`/chat/${chat.id}/agent/${customAgent.id}/edit`}
        method="post"
        ref={formRef}
        onChange={handleFormChange}
      >
        <CardHeader>
          <CardTitle>Agent Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6 mt-6">
          <Tabs defaultValue="local" className="w-full">
            <TabsList className="mb-4">
              <TabsTrigger value="local">Local</TabsTrigger>
              <TabsTrigger value="remote">Remote</TabsTrigger>
            </TabsList>

            <TabsContent value="local" className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  name="name"
                  placeholder="Agent name"
                  defaultValue={customAgent.name}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  name="description"
                  placeholder="Brief description of the agent"
                  defaultValue={customAgent.description}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="icon">Icon</Label>
                <Input
                  id="icon"
                  name="icon"
                  placeholder="Icon URL or path"
                  defaultValue={customAgent.icons?.light}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="darkIcon">Dark Icon</Label>
                <Input
                  id="darkIcon"
                  name="darkIcon"
                  placeholder="Dark mode icon URL or path"
                  defaultValue={customAgent.icons?.dark}
                />
              </div>

              {Object.keys(chat.agents || {}).length > 1 && (
                <div className="space-y-2">
                  <Label htmlFor="model">Model</Label>
                  <Select
                    value={selectedModel}
                    onValueChange={(value) => {
                      setSelectedModel(value);
                      setTimeout(handleFormChange, 0);
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select a model" />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(chat.agents ?? {}).map(([id, agent]) => (
                        <SelectItem key={id} value={id}>
                          {agent.name || id}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <input type="hidden" name="model" value={selectedModel} />
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="instructions">Instructions</Label>
                <Textarea
                  defaultValue={customAgent.instructions}
                  id="instructions"
                  name="instructions"
                  placeholder="Enter the system prompt"
                  className="min-h-[100px]"
                />
              </div>

              <div className="flex items-center space-x-2">
                <Switch
                  id="visibility"
                  checked={isPublic}
                  onCheckedChange={(checked) => {
                    setIsPublic(checked);
                    setTimeout(handleFormChange, 0);
                  }}
                />
                <Label htmlFor="visibility">Make this agent public</Label>
                <input
                  type="hidden"
                  name="isPublic"
                  value={isPublic ? "true" : "false"}
                />
              </div>
              <div className="space-y-2">
                <Label>MCP Server URLs</Label>
                {serverUrls.map((url, index) => (
                  <div key={index} className="flex gap-2 mb-2">
                    <Input
                      placeholder="Enter MCP server URL"
                      value={url}
                      name={`serverUrls[${index}]`}
                      onChange={(e) => handleUrlChange(index, e.target.value)}
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      className="px-2 h-10"
                      onClick={() => {
                        const newUrls = serverUrls.filter(
                          (_, i) => i !== index,
                        );
                        setServerUrls(newUrls.length ? newUrls : [""]);
                      }}
                    >
                      <MinusIcon className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
                {addNewUrlInput && (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setServerUrls([...serverUrls, ""])}
                  >
                    Add Another URL
                  </Button>
                )}
              </div>
            </TabsContent>

            <TabsContent value="remote" className="space-y-6">
              <div className="space-y-2">
                <Label htmlFor="remoteUrl">Remote URL</Label>
                <Input
                  id="remoteUrl"
                  name="remoteUrl"
                  placeholder="Enter remote URL"
                  defaultValue={customAgent.remoteUrl || ""}
                />
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
        <CardFooter className="flex justify-end space-x-2">
          <Link to={`/agent/${customAgent.id}`}>
            <Button variant="outline">Close</Button>
          </Link>
          <Button type="submit" disabled={!customAgent.id}>
            Save
          </Button>
        </CardFooter>
      </Form>
    </Card>
  );
}
