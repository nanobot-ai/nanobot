import { Suspense } from "react";
import { Await, useAsyncError } from "react-router";
import type { ChatsData } from "~/lib/nanobot";

function ErrorMsg() {
  const e = useAsyncError();
  return (
    <div>
      <h2>Error</h2>
      <pre>!!!{JSON.stringify(e)}</pre>
    </div>
  );
}

export default function Threads({ threads }: { threads: Promise<ChatsData> }) {
  return (
    <Suspense fallback={<div>Loading threads...</div>}>
      <Await resolve={threads} errorElement={<ErrorMsg />}>
        {(threads) => (
          <>
            <h1>Threads</h1>
            <pre>{JSON.stringify(threads, null, "  ")}</pre>
          </>
        )}
      </Await>
    </Suspense>
  );
}
