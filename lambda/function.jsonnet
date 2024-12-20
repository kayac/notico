local tfstate = std.native('tfstate');
local must_env = std.native('must_env');

{
  Architectures: [
    'arm64',
  ],
  EphemeralStorage: {
    Size: 512,
  },
  FunctionName: tfstate('aws_lambda_function.lambda.function_name'),
  Handler: 'bootstrap',
  LoggingConfig: {
    LogFormat: 'Text',
    LogGroup: tfstate('aws_cloudwatch_log_group.lambda.name'),
  },
  MemorySize: 128,
  Role: tfstate('aws_iam_role.lambda.arn'),
  Runtime: 'provided.al2023',
  SnapStart: {
    ApplyOn: 'None',
  },
  Tags: {
    ManagedBy: 'Terraform',
    Repo: 'kayac/notico',
    Service: 'notico',
  },
  Timeout: 3,
  TracingConfig: {
    Mode: 'PassThrough',
  },
  Environment: {
    Variables: {
      TZ: "Asia/Tokyo",
      SLACK_TOKEN: "secretfrom:aws_secretsmanager:/notico/prod.SLACK_TOKEN",
      SLACK_SIGNING_SECRET: "secretfrom:aws_secretsmanager:/notico/prod.SLACK_SIGNING_SECRET",
      DOMAIN: must_env('DOMAIN'),
      NOTICE_CHANNEL_ID: must_env('NOTICE_CHANNEL_ID'),
    },
  },
}
