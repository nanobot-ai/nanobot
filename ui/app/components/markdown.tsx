import { Link } from "react-router";
import { memo } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import { CodeBlock } from "./code-block";

const components: Partial<Components> = {
  // @ts-expect-error I dunno why this is needed, but it was just here
  code: CodeBlock,
  pre: ({ children }) => <>{children}</>,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ol: ({ node, children, ...props }) => {
    return (
      <ol className="list-decimal list-outside ml-4" {...props}>
        {children}
      </ol>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  li: ({ node, children, ...props }) => {
    return (
      <li className="py-1" {...props}>
        {children}
      </li>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ul: ({ node, children, ...props }) => {
    return (
      <ul className="list-decimal list-outside ml-4" {...props}>
        {children}
      </ul>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  strong: ({ node, children, ...props }) => {
    return (
      <span className="font-semibold" {...props}>
        {children}
      </span>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  a: ({ node, children, ...props }) => {
    return (
      props.href && (
        <Link
          className="text-blue-500 hover:underline"
          target="_blank"
          rel="noreferrer"
          to={props.href}
          {...props}
        >
          {children}
        </Link>
      )
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h1: ({ node, children, ...props }) => {
    return (
      <h1 className="text-3xl font-semibold mt-6 mb-2" {...props}>
        {children}
      </h1>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h2: ({ node, children, ...props }) => {
    return (
      <h2 className="text-2xl font-semibold mt-6 mb-2" {...props}>
        {children}
      </h2>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h3: ({ node, children, ...props }) => {
    return (
      <h3 className="text-xl font-semibold mt-6 mb-2" {...props}>
        {children}
      </h3>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h4: ({ node, children, ...props }) => {
    return (
      <h4 className="text-lg font-semibold mt-6 mb-2" {...props}>
        {children}
      </h4>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h5: ({ node, children, ...props }) => {
    return (
      <h5 className="text-base font-semibold mt-6 mb-2" {...props}>
        {children}
      </h5>
    );
  },
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  h6: ({ node, children, ...props }) => {
    return (
      <h6 className="text-sm font-semibold mt-6 mb-2" {...props}>
        {children}
      </h6>
    );
  },
};

const remarkPlugins = [remarkGfm];

const NonMemoizedMarkdown = ({ children }: { children: string }) => {
  return (
    <ReactMarkdown remarkPlugins={remarkPlugins} components={components}>
      {children}
    </ReactMarkdown>
  );
};

export const Markdown = memo(
  NonMemoizedMarkdown,
  (prevProps, nextProps) => prevProps.children === nextProps.children,
);
