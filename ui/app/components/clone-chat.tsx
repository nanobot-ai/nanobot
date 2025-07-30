import { Button } from "~/components/ui/button";
import { Form } from "react-router";

export default function CloneChat({ chatId }: { chatId: string }) {
  return (
    <Form method="post" action={`/chat/${chatId}/clone`} className="contents">
      <Button type="submit" className="mt-4 mx-auto">
        Continue Chat in New Thread
      </Button>
    </Form>
  );
}
