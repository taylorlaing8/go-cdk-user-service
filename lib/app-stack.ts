import * as cdk from 'aws-cdk-lib';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import * as backup from 'aws-cdk-lib/aws-backup';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as constructs from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as events from 'aws-cdk-lib/aws-events';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch';
import * as actions from 'aws-cdk-lib/aws-cloudwatch-actions';
import * as codedeploy from 'aws-cdk-lib/aws-codedeploy';
import * as ddb from 'aws-cdk-lib/aws-dynamodb';
import * as subscriptions from 'aws-cdk-lib/aws-sns-subscriptions';
import * as codeDeploy from 'aws-cdk-lib/aws-codedeploy';
import { Construct } from 'constructs';

export interface AppStackProps extends cdk.StackProps {
	stage: string;
	service: string;
	subscriptionEmail: string;
	authorizerFunctionArn: string;
	iso3166Code: string;
}

export class AppStack extends cdk.Stack {
	constructor(scope: constructs.Construct, id: string, props: AppStackProps) {
		super(scope, id, props);
		
		const snsTopic = new sns.Topic(this, 'SnsTopic', {
			topicName: `${this.stackName}-alarm`,
		});

		if (this.isProdStage(props.stage)) {
			snsTopic.addSubscription(
				new subscriptions.UrlSubscription(props.subscriptionEmail)
			);
		}

		// dynamoDb tables
		const usersTable = this.createUsersTable();

		const api = this.createApiGateway(props, snsTopic);

		if (this.isCicdStage(props.stage)) {
			const vault = new backup.BackupVault(this, 'BackupVault', {
				backupVaultName: this.stackName,
				removalPolicy: cdk.RemovalPolicy.DESTROY,
			});

			const plan = new backup.BackupPlan(this, 'BackupPlan', {
				backupPlanName: this.stackName,
				backupVault: vault,
			});

			plan.addSelection('BackupPlanSelection', {
				resources: [
					backup.BackupResource.fromDynamoDbTable(usersTable),
				],
			});

			plan.addRule(
				new backup.BackupPlanRule({
					startWindow: cdk.Duration.hours(1),
					completionWindow: cdk.Duration.hours(3),
					scheduleExpression: events.Schedule.cron({
						minute: '0',
						hour: '0',
						day: '1',
						month: '*',
						year: '*',
					}),
					moveToColdStorageAfter: cdk.Duration.days(30),
					deleteAfter: cdk.Duration.days(365),
				})
			);
		}

		// Identity Groups
		const createUser = this.createLambda(
			'CreateUser',
			'create-user',
			props,
			snsTopic,
			usersTable
		);
		usersTable.grantFullAccess(createUser);

		const updateUser = this.createLambda(
			'UpdateUser',
			'update-user',
			props,
			snsTopic,
			usersTable
		);
		usersTable.grantFullAccess(updateUser);

		const deleteUser = this.createLambda(
			'DeleteUser',
			'delete-user',
			props,
			snsTopic,
			usersTable
		);
		usersTable.grantFullAccess(deleteUser);

		const getUser = this.createLambda(
			'GetUser',
			'get-user',
			props,
			snsTopic,
			usersTable
		);
		usersTable.grantReadData(getUser);

		// Routes
		const v1 = api.root.addResource('v1');

		// Identity Groups
		const usersV1 = v1.addResource('users');
		usersV1.addMethod('POST', new apigateway.LambdaIntegration(createUser));

		const userIdV1 = usersV1.addResource('{userId}');
		userIdV1.addMethod('GET', new apigateway.LambdaIntegration(getUser));
		userIdV1.addMethod('PUT', new apigateway.LambdaIntegration(updateUser));
		userIdV1.addMethod(
			'DELETE',
			new apigateway.LambdaIntegration(deleteUser)
		);

		if (this.isCicdStage(props.stage)) {
			let environment = this.isCicdStage(props.stage)
				? props.stage
				: 'dev';

			new apigateway.CfnBasePathMapping(this, 'BasePathMapping', {
				domainName: `${environment}-api.classifind.app`,
				restApiId: api.restApiId,
				basePath: 'user',
				stage: api.deploymentStage.stageName,
			});
		}
	}

	private createAuthorizer(props: AppStackProps): apigateway.TokenAuthorizer {
		const authorizerFunction = lambda.Function.fromFunctionArn(
			this,
			'LambdaAuthorizer',
			props.authorizerFunctionArn
		);


		const authorizer = new apigateway.TokenAuthorizer(this, "Auth0Authorizer", {
			handler: authorizerFunction,
			authorizerName: "Auth0Authorizer",
			resultsCacheTtl: cdk.Duration.seconds(30),
		});

		return authorizer;
	}

	private createApiGateway(props: AppStackProps, snsTopic: sns.Topic): apigateway.RestApi {
		const authorizer = this.createAuthorizer(props);

		const apiResourcePolicy = new iam.PolicyDocument({
			statements: [
				new iam.PolicyStatement({
					effect: iam.Effect.ALLOW,
					actions: ['execute-api:Invoke'],
					principals: [new iam.AnyPrincipal()],
					resources: ['execute-api:/*'],
				}),
				// new iam.PolicyStatement({
				// 	effect: iam.Effect.DENY,
				// 	principals: [new iam.AnyPrincipal()],
				// 	actions: ['execute-api:Invoke'],
				// 	resources: ['execute-api:/*'],
				// 	conditions: {
				// 		StringNotEquals: {
				// 			'aws:SourceVpce': props.apiGatewayVpcEndpointId,
				// 		},
				// 	},
				// }),
			],
		});

		const accessLogs = new logs.LogGroup(this, 'AccessLogsLogGroup', {
			logGroupName: `/aws/api-gateway/${this.stackName}`,
			retention:
				props.stage === 'prod'
					? logs.RetentionDays.ONE_YEAR
					: logs.RetentionDays.ONE_WEEK,
			removalPolicy: cdk.RemovalPolicy.DESTROY,
		});

		const api = new apigateway.RestApi(this, 'api-gateway', {
			restApiName: this.stackName,
			endpointConfiguration: {
				// THIS SHOULD BECOME ENDPOINTYPE.EDGE WHEN YOU ARE READY TO EXPAND NATIONWIDE
				types: [apigateway.EndpointType.REGIONAL],
			},
			deployOptions: {
				stageName: 'LIVE',
				tracingEnabled: true,
				accessLogDestination: new apigateway.LogGroupLogDestination(
					accessLogs
				),
				accessLogFormat: apigateway.AccessLogFormat.custom(
					`{"requestTime":"${apigateway.AccessLogField.contextRequestTime()}","requestId":"${apigateway.AccessLogField.contextRequestId()}","httpMethod":"${apigateway.AccessLogField.contextHttpMethod()}","path":"${apigateway.AccessLogField.contextPath()}","resourcePath":"${apigateway.AccessLogField.contextResourcePath()}","status":${apigateway.AccessLogField.contextStatus()},"responseLatency":${apigateway.AccessLogField.contextResponseLatency()},"xrayTraceId":"${apigateway.AccessLogField.contextXrayTraceId()}","integrationLatency":"${apigateway.AccessLogField.contextIntegrationLatency()}","integrationStatus":"${apigateway.AccessLogField.contextIntegrationStatus()}","sourceIp":"${apigateway.AccessLogField.contextIdentitySourceIp()}","userAgent":"${apigateway.AccessLogField.contextIdentityUserAgent()}"}`
				),
			},
			defaultMethodOptions: {
				authorizer: authorizer,
			},
			policy: apiResourcePolicy,
			defaultCorsPreflightOptions: {
				allowMethods: apigateway.Cors.ALL_METHODS,
				allowOrigins: apigateway.Cors.ALL_ORIGINS,
				allowHeaders: apigateway.Cors.DEFAULT_HEADERS,
				maxAge: cdk.Duration.seconds(60),
			},
			cloudWatchRole: false,
		});

		api.addGatewayResponse('AccessDeniedGatewayResponse', {
			type: apigateway.ResponseType.ACCESS_DENIED,
			statusCode: '403',
			responseHeaders: {
				'Access-Control-Allow-Origin': "'*'",
				'Access-Control-Allow-Headers': "'*'",
				'Access-Control-Allow-Methods': "'*'",
				'Access-Control-Max-Age': "'86400'",
			},
			templates: {
				'application/json':
					'{ "ErrorMessage": "$context.authorizer.errorMessage", "ErrorCode": "$context.error.responseType", "Errors": [] }',
			},
		});

		api.addGatewayResponse('UnauthorizedGatewayResponse', {
			type: apigateway.ResponseType.UNAUTHORIZED,
			statusCode: '401',
			responseHeaders: {
				'Access-Control-Allow-Origin': "'*'",
				'Access-Control-Allow-Headers': "'*'",
				'Access-Control-Allow-Methods': "'*'",
				'Access-Control-Max-Age': "'86400'",
			},
			templates: {
				'application/json':
					'{ "ErrorMessage": "Unauthorized", "ErrorCode": "$context.error.responseType", "Errors": [] }',
			},
		});

		const apiErrors = new cloudwatch.Alarm(this, 'ApiErrors', {
			alarmDescription: '500 errors >= 5',
			metric: api.metricServerError({
				statistic: 'Sum',
				period: cdk.Duration.minutes(1),
			}),
			threshold: 5,
			evaluationPeriods: 1,
			actionsEnabled: true,
			comparisonOperator:
				cloudwatch.ComparisonOperator
					.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		});

		apiErrors.addAlarmAction(new actions.SnsAction(snsTopic));

		return api;
	}

	private createUsersTable(): ddb.Table {
		var table = new ddb.Table(this, 'UserTable', {
			tableName: `${this.stackName}-user`,
			billingMode: ddb.BillingMode.PAY_PER_REQUEST,
			partitionKey: {
				name: 'PK',
				type: ddb.AttributeType.STRING,
			},
			sortKey: {
				name: 'SK',
				type: ddb.AttributeType.STRING,
			},
			pointInTimeRecovery: true,
			removalPolicy: cdk.RemovalPolicy.DESTROY,
			encryption: ddb.TableEncryption.AWS_MANAGED,
			stream: ddb.StreamViewType.NEW_AND_OLD_IMAGES,
		});

		table.addGlobalSecondaryIndex({
			indexName: 'GSI1',
			partitionKey: {
				name: 'GSI1PK',
				type: ddb.AttributeType.STRING,
			},
			sortKey: {
				name: 'GSI1SK',
				type: ddb.AttributeType.STRING,
			},
		});

		table.addGlobalSecondaryIndex({
			indexName: 'GSI2',
			partitionKey: {
				name: 'GSI2PK',
				type: ddb.AttributeType.STRING,
			},
			sortKey: {
				name: 'GSI2SK',
				type: ddb.AttributeType.STRING,
			},
		});

		return table;
	}

	private createLambda(
		methodName: string,
		resourcePath: string,
		props: AppStackProps,
		snsTopic: sns.Topic,
		userTable: cdk.aws_dynamodb.Table
	): lambda.IFunction {
		const newLambda = new lambda.Function(this, methodName, {
			functionName: `${this.stackName}-${methodName}`,
			code: lambda.Code.fromAsset(`./dist/${resourcePath}/bootstrap.zip`),
			handler: 'bootstrap',
			runtime: lambda.Runtime.PROVIDED_AL2,
			architecture: lambda.Architecture.ARM_64,
			timeout: cdk.Duration.seconds(30),
			memorySize: 1024,
			environment: {
				SERVICE: props.service,
				STAGE: props.stage,
				USER_TABLE_NAME: userTable.tableName,
			},
			tracing: lambda.Tracing.ACTIVE,
			currentVersionOptions: {
				removalPolicy: cdk.RemovalPolicy.RETAIN,
			},
		});

		new logs.LogGroup(this, `${methodName}LogGroup`, {
			logGroupName: `/aws/lambda/${newLambda.functionName}`,
			retention:
				props.stage === 'prod'
					? logs.RetentionDays.ONE_YEAR
					: logs.RetentionDays.ONE_WEEK,
			removalPolicy: cdk.RemovalPolicy.DESTROY,
		});

		const newAlias = new lambda.Alias(this, `${methodName}Alias`, {
			aliasName: 'LIVE',
			version: newLambda.currentVersion,
		});

		const newErrors = new cloudwatch.Alarm(this, `${methodName}Errors`, {
			alarmName: `${this.stackName}-${methodName}-errors`,
			alarmDescription: 'The latest deployment errors > 0',
			metric: newAlias.metricErrors({
				statistic: 'Sum',
				period: cdk.Duration.minutes(1),
			}),
			threshold: 5,
			evaluationPeriods: 1,
			actionsEnabled: true,
			comparisonOperator:
				cloudwatch.ComparisonOperator
					.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		});

		newErrors.addAlarmAction(new actions.SnsAction(snsTopic));

		const lambdaDeploymentConfig = this.isProdStage(props.stage)
			? codeDeploy.LambdaDeploymentConfig.CANARY_10PERCENT_10MINUTES
			: codeDeploy.LambdaDeploymentConfig.ALL_AT_ONCE;

		new codedeploy.LambdaDeploymentGroup(
			this,
			`${methodName}DeploymentGroup`,
			{
				alias: newAlias,
				deploymentConfig: lambdaDeploymentConfig,
				alarms: [newErrors],
			}
		);

		return newAlias;
	}

	private isCicdStage(stage: string): boolean {
		return ['rd', 'dev', 'staging', 'prod'].includes(stage);
	}

	private isProdStage(stage: string): boolean {
		return 'prod' === stage;
	}
}
