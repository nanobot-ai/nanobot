import type { Workspace, WorkspaceFile } from '$lib/types';
import type { WorkspaceInstance } from '$lib/workspace.svelte';
import type { Task } from '../../routes/w/[workspaceId]/t/types';

export const sharedWorkspaces: Workspace[] = [
	{
		id: 'pytorch-onboarding',
		created: '2026-01-07',
		name: 'PyTorch Onboarding',
		color: '#2ddcec',
		order: 0
	},
	{
		id: 'aaif-onboarding',
		created: '2026-01-06',
		name: 'AAIF Onboarding',
		color: '#fdcc11',
		order: 0
	}
];

export const workspace: Workspace = {
	id: 'cncf-onboarding',
	created: '2026-01-02',
	name: 'CNCF Onboarding',
	color: '#000',
	order: 0
};

export const workspaceIds = [workspace.id, sharedWorkspaces[0].id, sharedWorkspaces[1].id];

export const taskIds = ['cncf_onboarding', 'pytorch_onboarding', 'aaif_onboarding'];

export const workspaceFiles: WorkspaceFile[] = [
	{
		name: '.nanobot/tasks/cncf_onboarding/TASK.md'
	}
];

export const sharedWorkspaceFiles: Record<string, WorkspaceFile[]> = {
	[sharedWorkspaces[0].id]: [
		{
			name: '.nanobot/tasks/pytorch_onboarding/TASK.md'
		}
	],
	[sharedWorkspaces[1].id]: [
		{
			name: '.nanobot/tasks/aaif_onboarding/TASK.md'
		}
	]
};
/**
"com.acornlabs.main/default-google-groups-f646a1f9x6fzp"
"com.acornlabs.main/default-slack-b73781ab"
"com.acornlabs.main/default-gmail-8a99d8be"
"com.acornlabs.main/default-github-391ae5a6"
 */

export const taskData: Record<string, Task> = {
	[taskIds[0]]: {
		name: 'Onboarding Workflow',
		description: '',
		inputs: [
			{
				name: 'CompanyName',
				description: 'The name of the company you want to add member to',
				default: 'Obot',
				id: 'eb3573cd-c6fb-4dda-beda-a92623e90fb4'
			}
		],
		steps: [
			{
				id: 'TASK.md',
				name: 'Add Member To Google Sheets',
				description: 'This task will add the member to the google sheet',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). Notes or special instructions are typically found in the Account’s Description field. To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Get the Demo Workflow LF Google Sheet. Read the first few rows to understand the sheet and formats used in each column.\n3. Follow the formatting and style, add a new row to the google sheet for the member $CompanyName based on the information we got previously in Salesforce. The join date should be the closed won date.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-sheets-68166c0a',
						title: 'Google Sheets',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-sheets-68166c0a'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_1.md',
				name: 'Add New Member Contacts to Google Groups',
				description:
					'This will add the contacts of our new members to the appropriate Google Groups',
				content:
					'1. Get the account record for $CompanyName  in Salesforce. You also need to get all related Contacts including roles and emails.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-groups-f646a1f9x6fzp',
						title: 'Google Groups',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-groups-f646a1f9x6fzp'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_2.md',
				name: 'Add Member Contacts To Slack',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.\n2. List all channels including private ones.\n3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.\n4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.\n5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.\n6. Report back who is a member of slack already.',
				tools: [
					{
						name: 'com.acornlabs.main/default-slack-b73781ab',
						title: 'Slack',
						url: 'https://main.acornlabs.com/mcp-connect/default-slack-b73781ab'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_3.md',
				name: 'Add Logo to Site',
				description: 'Create a github PR to add the logo to the site',
				content:
					'1. Get the account record for $CompanyName in Salesforce. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Create a branch in the repo cloudnautique/obot-mcpserver-examples called add-$CompanyName-logo.\n3. Create a file in the workspace called logo.txt and write a story about a robot in markdown.\n4. Add the file to the assets/img directory called $CompanyName-logo.txt, and create a PR back into the main branch.',
				tools: [
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					},
					{
						name: 'com.acornlabs.main/default-github-391ae5a6',
						title: 'GitHub',
						url: 'https://main.acornlabs.com/mcp-connect/default-github-391ae5a6'
					}
				]
			},
			{
				id: 'STEP_4.md',
				name: 'Send Welcome Email',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get the contacts and their email addresses. Also get the most recent opportunity to determine the membership.\n2. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info. \n\n   ```Markdown\n   # CNCF Onboarding Completion \n\n   **Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!\n\n   ---\n\n   Dear {{FirstName}} {{LastName}},\n\n   Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!\n\n   We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.\n\n   Congrats!\n\n   Dir. CNCF Onboarding Agent\n\n   ```\n4. Send the drafted email.',
				tools: [
					{
						name: 'com.acornlabs.main/default-gmail-8a99d8be',
						title: 'Gmail',
						url: 'https://main.acornlabs.com/mcp-connect/default-gmail-8a99d8be'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			}
		]
	},

	[taskIds[1]]: {
		name: 'PyTorch Onboarding',
		description: '',
		inputs: [
			{
				name: 'ProductName',
				description: 'The name of the product',
				id: 'eb3573cd-c6fb-4dda-beda-a92623e90fb4'
			}
		],
		steps: [
			{
				id: 'TASK.md',
				name: 'Add Member To Google Sheets',
				description: 'This task will add the member to the google sheet',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). Notes or special instructions are typically found in the Account’s Description field. To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Get the Demo Workflow LF Google Sheet. Read the first few rows to understand the sheet and formats used in each column.\n3. Follow the formatting and style, add a new row to the google sheet for the member $CompanyName based on the information we got previously in Salesforce. The join date should be the closed won date.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-sheets-68166c0a',
						title: 'Google Sheets',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-sheets-68166c0a'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_1.md',
				name: 'Add New Member Contacts to Google Groups',
				description:
					'This will add the contacts of our new members to the appropriate Google Groups',
				content:
					'1. Get the account record for $CompanyName  in Salesforce. You also need to get all related Contacts including roles and emails.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-groups-f646a1f9x6fzp',
						title: 'Google Groups',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-groups-f646a1f9x6fzp'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_2.md',
				name: 'Add Member Contacts To Slack',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.\n2. List all channels including private ones.\n3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.\n4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.\n5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.\n6. Report back who is a member of slack already.',
				tools: [
					{
						name: 'com.acornlabs.main/default-slack-b73781ab',
						title: 'Slack',
						url: 'https://main.acornlabs.com/mcp-connect/default-slack-b73781ab'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_3.md',
				name: 'Add Logo to Site',
				description: 'Create a github PR to add the logo to the site',
				content:
					'1. Get the account record for $CompanyName in Salesforce. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Create a branch in the repo cloudnautique/obot-mcpserver-examples called add-$CompanyName-logo.\n3. Create a file in the workspace called logo.txt and write a story about a robot in markdown.\n4. Add the file to the assets/img directory called $CompanyName-logo.txt, and create a PR back into the main branch.',
				tools: [
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					},
					{
						name: 'com.acornlabs.main/default-github-391ae5a6',
						title: 'GitHub',
						url: 'https://main.acornlabs.com/mcp-connect/default-github-391ae5a6'
					}
				]
			},
			{
				id: 'STEP_4.md',
				name: 'Send Welcome Email',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get the contacts and their email addresses. Also get the most recent opportunity to determine the membership.\n2. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info. \n\n   ```Markdown\n   # CNCF Onboarding Completion \n\n   **Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!\n\n   ---\n\n   Dear {{FirstName}} {{LastName}},\n\n   Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!\n\n   We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.\n\n   Congrats!\n\n   Dir. CNCF Onboarding Agent\n\n   ```\n4. Send the drafted email.',
				tools: [
					{
						name: 'com.acornlabs.main/default-gmail-8a99d8be',
						title: 'Gmail',
						url: 'https://main.acornlabs.com/mcp-connect/default-gmail-8a99d8be'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			}
		]
	},
	[taskIds[2]]: {
		name: 'AAIF Onboarding',
		description: '',
		inputs: [
			{
				name: 'CompanyName',
				description: 'The name of the company you want to add member to',
				id: 'eb3573cd-c6fb-4dda-beda-a92623e90fb4'
			},
			{
				name: 'EmployeeName',
				description: 'The name of the employee you want to add',
				id: 'eb3573cd-c6fb-4dda-beda-a92623e90fb5'
			}
		],
		steps: [
			{
				id: 'TASK.md',
				name: 'Add Member To Google Sheets',
				description: 'This task will add the member to the google sheet',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). Notes or special instructions are typically found in the Account’s Description field. To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Get the Demo Workflow LF Google Sheet. Read the first few rows to understand the sheet and formats used in each column.\n3. Follow the formatting and style, add a new row to the google sheet for the member $CompanyName based on the information we got previously in Salesforce. The join date should be the closed won date.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-sheets-68166c0a',
						title: 'Google Sheets',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-sheets-68166c0a'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_1.md',
				name: 'Add New Member Contacts to Google Groups',
				description:
					'This will add the contacts of our new members to the appropriate Google Groups',
				content:
					'1. Get the account record for $CompanyName  in Salesforce. You also need to get all related Contacts including roles and emails.',
				tools: [
					{
						name: 'com.acornlabs.main/default-google-groups-f646a1f9x6fzp',
						title: 'Google Groups',
						url: 'https://main.acornlabs.com/mcp-connect/default-google-groups-f646a1f9x6fzp'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_2.md',
				name: 'Add Member Contacts To Slack',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.\n2. List all channels including private ones.\n3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.\n4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.\n5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.\n6. Report back who is a member of slack already.',
				tools: [
					{
						name: 'com.acornlabs.main/default-slack-b73781ab',
						title: 'Slack',
						url: 'https://main.acornlabs.com/mcp-connect/default-slack-b73781ab'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			},
			{
				id: 'STEP_3.md',
				name: 'Add Logo to Site',
				description: 'Create a github PR to add the logo to the site',
				content:
					'1. Get the account record for $CompanyName in Salesforce. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Create a branch in the repo cloudnautique/obot-mcpserver-examples called add-$CompanyName-logo.\n3. Create a file in the workspace called logo.txt and write a story about a robot in markdown.\n4. Add the file to the assets/img directory called $CompanyName-logo.txt, and create a PR back into the main branch.',
				tools: [
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					},
					{
						name: 'com.acornlabs.main/default-github-391ae5a6',
						title: 'GitHub',
						url: 'https://main.acornlabs.com/mcp-connect/default-github-391ae5a6'
					}
				]
			},
			{
				id: 'STEP_4.md',
				name: 'Send Welcome Email',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get the contacts and their email addresses. Also get the most recent opportunity to determine the membership.\n2. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info. \n\n   ```Markdown\n   # CNCF Onboarding Completion \n\n   **Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!\n\n   ---\n\n   Dear {{FirstName}} {{LastName}},\n\n   Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!\n\n   We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.\n\n   Congrats!\n\n   Dir. CNCF Onboarding Agent\n\n   ```\n4. Send the drafted email.',
				tools: [
					{
						name: 'com.acornlabs.main/default-gmail-8a99d8be',
						title: 'Gmail',
						url: 'https://main.acornlabs.com/mcp-connect/default-gmail-8a99d8be'
					},
					{
						name: 'com.acornlabs.main/default-salesforce-f032ecc7',
						title: 'Salesforce',
						url: 'https://main.acornlabs.com/mcp-connect/default-salesforce-f032ecc7'
					}
				]
			}
		]
	}
};

export const stepSummaries: Record<string, { step: string; summary: string }[]> = {
	[taskIds[0]]: [
		{
			step: 'Add Member To Google Sheets',
			summary:
				'John Doe, Jane Doe, Mark Smith, Emily Johnson, David Lee were added to the google sheet.'
		},
		{
			step: 'Add New Member Contacts to Google Groups',
			summary:
				'John Doe, Jane Doe, Mark Smith, Emily Johnson, David Lee were added to the Marketing & Business Proposals groups.'
		},
		{
			step: 'Add Member Contacts To Slack',
			summary:
				'John Doe, Jane Doe, Mark Smith, Emily Johnson, David Lee were added to the private marketing channel, business-owners channel, and technical-leads channel.'
		},
		{ step: 'Add Logo to Site', summary: 'A PR was created to add the logo to the site.' },
		{
			step: 'Send Welcome Email',
			summary: 'Welcome email sent to John Doe, Jane Doe, Mark Smith, Emily Johnson, David Lee.'
		}
	],
	[taskIds[1]]: [
		{
			step: 'Search popular sites for the product',
			summary:
				'Product found on Amazon for $100.00, Walmart for $101.00, Best Buy for $102.00, Target for $103.00, and eBay for $104.00. The most affordable product is Amazon for $100.00; an email was sent to the user with the product details.'
		}
	],
	[taskIds[2]]: [
		{ step: 'Add Member To Google Sheets', summary: 'John Doe was added to Google Sheets' },
		{
			step: 'Add Member Contact Information to Google Groups',
			summary: 'John Doe was added to the Marketing & Business Proposals groups'
		},
		{
			step: 'Add Member To Slack',
			summary:
				'John Doe was added to the private marketing channel, business-owners channel, and technical-leads channel'
		},
		{ step: 'Send Welcome Email', summary: 'The welcome email was sent to John Doe.' }
	]
};

export const workspaceInstances: Record<string, WorkspaceInstance> = {
	[workspace.id]: {
		files: workspaceFiles,
		sessions: [],
		loading: false
	} as unknown as WorkspaceInstance,
	[sharedWorkspaces[0].id]: {
		files: sharedWorkspaceFiles[sharedWorkspaces[0].id],
		sessions: [],
		loading: false
	} as unknown as WorkspaceInstance,
	[sharedWorkspaces[1].id]: {
		files: sharedWorkspaceFiles[sharedWorkspaces[1].id],
		sessions: [],
		loading: false
	} as unknown as WorkspaceInstance
};

export const workspacePermissions: Record<string, string[]> = {
	'pytorch-onboarding': ['read', 'write', 'execute'],
	'aaif-onboarding': ['execute']
};

export const tasks = [
	{
		id: 'cncf_onboarding',
		name: 'Onboarding',
		created: '2026-01-02',
		workspace: 'CNCF Onboarding'
	},
	{
		id: 'pytorch_onboarding',
		name: 'Onboarding',
		created: '2026-01-01',
		workspace: 'PyTorch Onboarding'
	},
	{ id: 'aaif_onboarding', name: 'Onboarding', created: '2026-01-01', workspace: 'AAIF Onboarding' }
];

export const files = [
	{
		id: 'cncf_onboarding.xlsx',
		name: 'Users to Onboard.xlsx',
		created: '2026-01-03',
		workspace: 'CNCF Onboarding'
	},
	{
		id: 'pytorch_onboarding.doc',
		name: 'Onboarding Process.doc ',
		created: '2026-01-02',
		workspace: 'PyTorch Onboarding'
	}
];

export const taskRuns = [
	{
		id: '1',
		created: '2026-01-03 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 7000
	},
	{
		id: '2',
		created: '2026-01-02 10:00:00',
		task: 'PyTorch Onboarding',
		averageCompletionTime: '10.1m',
		user: 'John Doe',
		workspace: 'PyTorch Onboarding',
		tokensUsed: 8500
	},
	{
		id: '3',
		created: '2026-01-02 10:00:00',
		task: 'AAIF Onboarding',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'AAIF Onboarding',
		tokensUsed: 8000
	},
	{
		id: '4',
		created: '2026-01-01 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '11m',
		user: 'Jane Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 9000
	},
	{
		id: '5',
		created: '2026-01-01 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 10000
	},
	{
		id: '6',
		created: '2026-01-01 10:00:00',
		task: 'PyTorch Onboarding',
		averageCompletionTime: '6.5m',
		user: 'John Doe',
		workspace: 'PyTorch Onboarding',
		tokensUsed: 11500
	},
	{
		id: '7',
		created: '2026-01-01 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 12000
	},
	{
		id: '8',
		created: '2026-01-01 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 13000
	},
	{
		id: '9',
		created: '2026-01-01 10:00:00',
		task: 'CNCF Onboarding',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'CNCF Onboarding',
		tokensUsed: 14000
	},
	{
		id: '10',
		created: '2026-01-01 10:00:00',
		task: 'PyTorch Onboarding',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'PyTorch Onboarding',
		tokensUsed: 15500
	}
];

export const users = [
	{
		id: '1',
		name: 'John Doe',
		email: 'john.doe@example.com'
	},
	{
		id: '2',
		name: 'Jane Doe',
		email: 'jane.doe@example.com'
	},
	{
		id: '3',
		name: 'Mark Smith',
		email: 'mark.smith@example.com'
	},
	{
		id: '4',
		name: 'Emily Johnson',
		email: 'emily.johnson@example.com'
	},
	{
		id: '5',
		name: 'David Lee',
		email: 'david.lee@example.com'
	}
];
