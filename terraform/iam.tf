resource "aws_iam_role" "lambda" {
  name               = "notico-lambda"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  path               = "/service-role/"
}

resource "aws_iam_role_policy_attachment" "lambda" {
  policy_arn = aws_iam_policy.lambda.arn
  role       = aws_iam_role.lambda.name
}

resource "aws_iam_policy" "lambda" {
  name   = "notico-lambda"
  policy = data.aws_iam_policy_document.lambda.json
  path   = "/service-role/"
}

data "aws_iam_policy_document" "lambda" {
  statement {
    actions   = ["logs:CreateLogStream", "logs:PutLogEvents"]
    effect    = "Allow"
    resources = ["${aws_cloudwatch_log_group.lambda.arn}:*"]
  }
  statement {
    actions   = ["secretsmanager:BatchGetSecretValue"]
    effect    = "Allow"
    resources = ["*"]
  }
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    effect    = "Allow"
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

