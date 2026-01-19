export const load = ({ params }) => {
	const { workspaceId, taskId } = params;
	return {
		workspaceId,
        taskId
	};
};
