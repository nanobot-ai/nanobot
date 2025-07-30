import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Form } from "react-router";

export function NewCustomAgent({
  chatId,
  onOpenChange,
  open = false,
}: {
  chatId: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <Form
          method="post"
          action={`/chat/${chatId}`}
          className="contents"
          onSubmit={(e) => {
            onOpenChange?.(false);
          }}
        >
          <DialogHeader>
            <DialogTitle>New Custom Agent</DialogTitle>
            <DialogDescription>
              Name your new agent. You can change it later.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4">
            <div className="grid gap-3">
              <Label htmlFor="name-1">Agent Name</Label>
              <Input id="name-1" name="newAgentName" defaultValue="" />
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button type="submit">Create</Button>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
