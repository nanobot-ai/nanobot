import { type ChatMessage, type Chat } from '$lib/types';

export const threads: Chat[] = [
	{
		id: '1',
		title: 'Project Setup Help',
		created: new Date(Date.now() - 1000 * 60 * 5).toISOString() // 5 minutes ago
	},
	{
		id: '2',
		title: 'TypeScript Configuration',
		created: new Date(Date.now() - 1000 * 60 * 30).toISOString() // 30 minutes ago
	},
	{
		id: '3',
		title: 'Database Schema Design',
		created: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString() // 2 hours ago
	},
	{
		id: '4',
		title: 'API Rate Limiting',
		created: new Date(Date.now() - 1000 * 60 * 60 * 6).toISOString() // 6 hours ago
	},
	{
		id: '5',
		title: 'Deployment Issues',
		created: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString() // 1 day ago
	},
	{
		id: '6',
		title: 'Performance Optimization',
		created: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2).toISOString() // 2 days ago
	}
];

export const messages: ChatMessage[] = [
	{
		id: '1',
		role: 'user',
		created: new Date(Date.now() - 10000).toISOString(),
		items: [
			{
				id: 'item-1',
				type: 'text',
				text: "Hello! I'm looking for help with my project setup."
			}
		]
	},
	{
		id: '2',
		role: 'assistant',
		created: new Date(Date.now() - 8000).toISOString(),
		items: [
			{
				id: 'item-2',
				type: 'text',
				text: "I'd be happy to help you with your project setup! What specific aspects are you working on?"
			}
		]
	},
	{
		id: '3',
		role: 'user',
		created: new Date(Date.now() - 7000).toISOString(),
		items: [
			{
				id: 'item-3a',
				type: 'text',
				text: "Here's an image of my current setup:"
			},
			{
				id: 'item-3b',
				type: 'image',
				data:
					'/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDADIiJSwlHzIsKSw4NTI7S31RS0VFS5ltc1p9tZ++u7Kf\n' +
					'r6zI4f/zyNT/16yv+v/9////////wfD/////////////2wBDATU4OEtCS5NRUZP/zq/O////////\n' +
					'////////////////////////////////////////////////////////////wAARCAAYAEADAREA\n' +
					'AhEBAxEB/8QAGQAAAgMBAAAAAAAAAAAAAAAAAQMAAgQF/8QAJRABAAIBBAEEAgMAAAAAAAAAAQIR\n' +
					'AAMSITEEEyJBgTORUWFx/8QAFAEBAAAAAAAAAAAAAAAAAAAAAP/EABQRAQAAAAAAAAAAAAAAAAAA\n' +
					'AAD/2gAMAwEAAhEDEQA/AOgM52xQDrjvAV5Xv0vfKUALlTQfeBm0HThMNHXkL0Lw/swN5qgA8yT4\n' +
					'MCS1OEOJV8mBz9Z05yfW8iSx7p4j+jA1aD6Wj7ZMzstsfvAas4UyRHvjrAkC9KhpLMClQntlqFc2\n' +
					'X1gUj4viwVObKrddH9YDoHvuujAEuNV+bLwFS8XxdSr+Cq3Vf+4F5RgQl6ZR2p1eAzU/HX80YBYy\n' +
					'JLCuexwJCO2O1bwCRidAfWBSctswbI12GAJT3yiwFR7+MBjGK2g/WAJR3FdF84E2rK5VR0YH/9k=',
				mimeType: 'image/png'
			}
		]
	},
	{
		id: '4',
		role: 'assistant',
		created: new Date(Date.now() - 6000).toISOString(),
		items: [
			{
				id: 'item-4a',
				type: 'reasoning',
				summary: [
					{ text: 'The user has shared an image of their setup and is asking for help.' },
					{ text: 'I should provide comprehensive guidance about project configuration.' }
				]
			},
			{
				id: 'item-4b',
				type: 'text',
				text: "Great! For development environment setup, you'll want to install the dependencies first using `pnpm install`. Then you can run `pnpm dev` to start the development server."
			}
		]
	},
	{
		id: '5',
		role: 'assistant',
		created: new Date(Date.now() - 5000).toISOString(),
		items: [
			{
				id: 'item-5a',
				type: 'text',
				text: 'Here is a sample SvelteKit component with **markdown formatting** and `inline code`:\n\n```typescript\nimport { onMount } from "svelte";\nimport type { PageData } from "./$types";\n\ninterface User {\n  id: number;\n  name: string;\n  email: string;\n  isActive: boolean;\n}\n\nlet users: User[] = $state([]);\nlet loading = $state(true);\n\nonMount(() => {\n  fetchUsers();\n});\n```\n\nThis shows:\n- **TypeScript interfaces**\n- **Svelte 5 runes** for state management'
			},
			{
				id: 'item-5b',
				type: 'resource_link',
				name: 'SvelteKit Documentation',
				description: 'Complete guide to SvelteKit framework',
				uri: 'https://kit.svelte.dev/docs'
			}
		]
	},
	{
		id: '6',
		role: 'assistant',
		created: new Date(Date.now() - 4000).toISOString(),
		items: [
			{
				id: 'item-6a',
				type: 'text',
				text: 'Let me check the current status of your database to see the user table structure.'
			},
			{
				id: 'item-6b',
				type: 'tool',
				name: 'execute_sql',
				arguments:
					'{"query": "SELECT table_name, column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = \'users\' ORDER BY ordinal_position;", "database": "production"}',
				output: {
					content: [
						{
							id: 'tool-output-6b',
							type: 'text',
							text: '```\ntable_name | column_name | data_type | is_nullable\nusers      | id          | uuid      | NO\nusers      | email       | varchar   | NO\nusers      | name        | varchar   | YES\nusers      | created_at  | timestamp | NO\nusers      | updated_at  | timestamp | NO\nusers      | is_active   | boolean   | NO\n\n(6 rows)\n```'
						}
					]
				}
			}
		]
	},
	{
		id: '7',
		role: 'user',
		created: new Date(Date.now() - 3000).toISOString(),
		items: [
			{
				id: 'item-7a',
				type: 'text',
				text: 'Can you also check this resource file?'
			},
			{
				id: 'item-7b',
				type: 'resource',
				resource: {
					uri: 'file:///config/database.yml',
					mimeType: 'text/yaml',
					text: 'production:\n  host: localhost\n  database: myapp\n  username: postgres\n  password: secret\n  port: 5432'
				}
			}
		]
	},
	{
		id: '8',
		role: 'assistant',
		created: new Date(Date.now() - 2000).toISOString(),
		items: [
			{
				id: 'item-8a',
				type: 'text',
				text: 'Now let me fetch the user data from the API to show you the current users.'
			},
			{
				id: 'item-8b',
				type: 'tool',
				name: 'api_call',
				arguments:
					'{"url": "https://api.example.com/users", "method": "GET", "headers": {"Authorization": "Bearer token123"}, "params": {"limit": 5, "active": true}}',
				output: {
					content: [
						{
							id: 'tool-output-8b',
							type: 'text',
							text: '```json\n{\n  "users": [\n    {\n      "id": 1,\n      "name": "John Doe",\n      "email": "john@example.com",\n      "is_active": true,\n      "created_at": "2024-01-15T10:30:00Z"\n    },\n    {\n      "id": 2,\n      "name": "Jane Smith",\n      "email": "jane@example.com",\n      "is_active": true,\n      "created_at": "2024-01-16T14:22:00Z"\n    }\n  ],\n  "total": 2,\n  "success": true\n}\n```'
						}
					]
				}
			}
		]
	}
];
