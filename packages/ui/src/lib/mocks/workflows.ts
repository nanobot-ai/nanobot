import type { Workspace, WorkspaceFile } from '$lib/types';
import type { WorkspaceInstance } from '$lib/workspace.svelte';
import type { Task } from '../../routes/w/[workspaceId]/t/types';

export const sharedWorkspaces: Workspace[] = [
	{
		id: 'mock-matcha-latte',
		created: '2026-01-07',
		name: 'Matcha Latte',
		color: '#2ddcec',
		order: 0
	},
	{
		id: 'mock-pumpkin-spice',
		created: '2026-01-06',
		name: 'Pumpkin Spice',
		color: '#fdcc11',
		order: 0
	}
];

export const workspace: Workspace = {
	id: 'mock-onboarding-support',
	created: '2026-01-02',
	name: 'Onboarding Support',
	color: '#000',
	order: 0
};

export const workspaceIds = [workspace.id, sharedWorkspaces[0].id, sharedWorkspaces[1].id];

export const taskIds = [
	'onboarding_workflow',
	'find_most_affordable_product',
	'modified_onboarding_workflow'
];

export const workspaceFiles: WorkspaceFile[] = [
	{
		name: '.nanobot/tasks/onboarding_workflow/TASK.md'
	}
];

export const sharedWorkspaceFiles: Record<string, WorkspaceFile[]> = {
	[sharedWorkspaces[0].id]: [
		{
			name: '.nanobot/tasks/find_most_affordable_product/TASK.md'
		}
	],
	[sharedWorkspaces[1].id]: [
		{
			name: '.nanobot/tasks/modified_onboarding_workflow/TASK.md'
		}
	]
};

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
					'com.acornlabs.main/default-google-sheets-68166c0a',
					'com.acornlabs.main/default-salesforce-f032ecc7'
				]
			},
			{
				id: 'STEP_1.md',
				name: 'Add New Member Contacts to Google Groups',
				description:
					'This will add the contacts of our new members to the appropriate Google Groups',
				content:
					'1. Get the account record for $CompanyName  in Salesforce. You also need to get all related Contacts including roles and emails.',
				tools: []
			},
			{
				id: 'STEP_2.md',
				name: 'Add Member Contacts To Slack',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get all related Contacts including roles and emails.\n2. List all channels including private ones.\n3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.\n4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.\n5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.\n6. Report back who is a member of slack already.',
				tools: []
			},
			{
				id: 'STEP_3.md',
				name: 'Add Logo to Site',
				description: 'Create a github PR to add the logo to the site',
				content:
					'1. Get the account record for $CompanyName in Salesforce. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Create a branch in the repo cloudnautique/obot-mcpserver-examples called add-$CompanyName-logo.\n3. Create a file in the workspace called logo.txt and write a story about a robot in markdown.\n4. Add the file to the assets/img directory called $CompanyName-logo.txt, and create a PR back into the main branch.',
				tools: []
			},
			{
				id: 'STEP_4.md',
				name: 'Send Welcome Email',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. You also need to get the contacts and their email addresses. Also get the most recent opportunity to determine the membership.\n2. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info. \n\n   ```Markdown\n   # CNCF Onboarding Completion \n\n   **Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!\n\n   ---\n\n   Dear {{FirstName}} {{LastName}},\n\n   Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!\n\n   We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.\n\n   Congrats!\n\n   Dir. CNCF Onboarding Agent\n\n   ```\n4. Send the drafted email.',
				tools: []
			}
		]
	},

	[taskIds[1]]: {
		name: 'Find Most Affortable Product',
		description: '',
		inputs: [
			{
				name: 'ProductName',
				description: 'The name of the product',
				default: 'Obot',
				id: 'eb3573cd-c6fb-4dda-beda-a92623e90fb4'
			}
		],
		steps: [
			{
				id: 'TASK.md',
				name: 'Search popular sites for the product',
				description: '',
				tools: [],
				content:
					'1. Search the following sites for the product $ProductName: Amazon, Walmart, Best Buy, Target, and eBay.'
			}
		]
	},
	[taskIds[2]]: {
		name: 'Modified Onboarding Workflow',
		description: '',
		inputs: [
			{
				name: 'CompanyName',
				description: 'The name of the company you want to add member to',
				default: 'Obot',
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
					'1. Get the account record for $CompanyName in Salesforce. Specifically get $EmployeeName contact information including roles and emails. Search all opportunities, look for the most recent closed won opportunity for membership level. Get the documents. Documents may be stored as classic Attachments (in the Attachment object, linked by ParentId) or as Salesforce Files (in ContentDocument, linked to the Account via ContentDocumentLink and LinkedEntityId). Notes or special instructions are typically found in the Account’s Description field. To review everything created for a company, look for these related records and fields using the Account’s Id.\n2. Get the Demo Workflow LF Google Sheet. Read the first few rows to understand the sheet and formats used in each column.\n3. Follow the formatting and style, add a new row to the google sheet for the member $CompanyName based on the information we got previously in Salesforce. The join date should be the closed won date.',
				tools: [
					'com.acornlabs.main/default-google-sheets-68166c0a',
					'com.acornlabs.main/default-salesforce-f032ecc7'
				]
			},
			{
				id: 'STEP_1.md',
				name: 'Add Member Contact Information to Google Groups',
				description:
					'Add the contact information of $EmployeeName to the appropriate Google Groups',
				content:
					'1. Get the account record for $CompanyName  in Salesforce. You also need to get all related Contacts including roles and emails.',
				tools: []
			},
			{
				id: 'STEP_2.md',
				name: 'Add Member To Slack',
				description: '',
				content:
					'1. Get the account record for $CompanyName in Salesforce. Get the contact information for $EmployeeName including roles and emails.\n2. List all channels including private ones.\n3. Search for the marketing contacts in the slack workspace to see if they are members. If they are in the workspace add them to the private marketing channel.\n4. Search for the business owner contacts in the slack workspace to see if they are members. If they are in the workspace, add them to the private business-owners channel.\n5. Search for the technical contacts in the slack workspace to see if they are members. If they are members of the workspace, add them to the private technical-leads channel.\n6. Report back who is a member of slack already.',
				tools: []
			},
			{
				id: 'STEP_4.md',
				name: 'Send Welcome Email',
				description: '',
				content:
					'1. Using gmail tools, create a draft email using the business owner contact, account, and opportunity info. \n\n   ```Markdown\n   # CNCF Onboarding Completion \n\n   **Subject:** Welcome to the Cloud Native Computing Foundation (CNCF)!\n\n   ---\n\n   Dear {{FirstName}} {{LastName}},\n\n   Congratulations, and welcome to the **Cloud Native Computing Foundation (CNCF)** community!\n\n   We’re pleased to let you know that all onboarding steps for **{{CompanyName}}** have been successfully completed. Your organization is now fully set up as a {{membership level}} member and ready to take advantage of CNCF programs, resources, and community benefits.\n\n   Congrats!\n\n   Dir. CNCF Onboarding Agent\n\n   ```\n4. Send the drafted email.',
				tools: []
			}
		]
	}
};

export const workspaceInstance = {
	files: workspaceFiles,
	sessions: [],
	loading: false
} as unknown as WorkspaceInstance;

export const workspacePermissions: Record<string, string[]> = {
	'mock-matcha-latte': ['read', 'write', 'execute'],
	'mock-pumpkin-spice': ['execute']
};

export const tasks = [
	{ id: '1', name: 'Onboarding Workflow', created: '2026-01-02', workspace: 'Adorable Akita' },
	{ id: '2', name: 'Customer Support', created: '2026-01-01', workspace: 'Adorable Akita' },
	{ id: '3', name: 'Marketing Campaign', created: '2026-01-01', workspace: 'Caramel Cookie' }
];

export const files = [
	{ id: '1', name: 'Example.pdf', created: '2026-01-03', workspace: 'Adorable Akita' },
	{ id: '2', name: 'Example.docx', created: '2026-01-02', workspace: 'Caramel Cookie' },
	{ id: '3', name: 'Example.xlsx', created: '2026-01-01', workspace: 'Adorable Akita' }
];

export const taskRuns = [
	{
		id: '1',
		created: '2026-01-03 10:00:00',
		task: 'Onboarding Workflow',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 7000
	},
	{
		id: '2',
		created: '2026-01-02 10:00:00',
		task: 'Customer Support',
		averageCompletionTime: '10.1m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 8500
	},
	{
		id: '3',
		created: '2026-01-02 10:00:00',
		task: 'Marketing Campaign',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'Caramel Cookie',
		tokensUsed: 8000
	},
	{
		id: '4',
		created: '2026-01-01 10:00:00',
		task: 'Product Launch',
		averageCompletionTime: '11m',
		user: 'Jane Doe',
		workspace: 'Caramel Cookie',
		tokensUsed: 9000
	},
	{
		id: '5',
		created: '2026-01-01 10:00:00',
		task: 'Sales Pipeline',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 10000
	},
	{
		id: '6',
		created: '2026-01-01 10:00:00',
		task: 'Customer Support',
		averageCompletionTime: '6.5m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 11500
	},
	{
		id: '7',
		created: '2026-01-01 10:00:00',
		task: 'Marketing Campaign',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'Caramel Cookie',
		tokensUsed: 12000
	},
	{
		id: '8',
		created: '2026-01-01 10:00:00',
		task: 'Product Launch',
		averageCompletionTime: '10m',
		user: 'Jane Doe',
		workspace: 'Caramel Cookie',
		tokensUsed: 13000
	},
	{
		id: '9',
		created: '2026-01-01 10:00:00',
		task: 'Sales Pipeline',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 14000
	},
	{
		id: '10',
		created: '2026-01-01 10:00:00',
		task: 'Customer Support',
		averageCompletionTime: '10m',
		user: 'John Doe',
		workspace: 'Adorable Akita',
		tokensUsed: 15500
	}
];
