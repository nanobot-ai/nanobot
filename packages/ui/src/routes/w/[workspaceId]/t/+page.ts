export const load = ({ params }) => {
	const { workspaceId } = params;
	return {
		inverse: true,
		workspaceId
	};
};
