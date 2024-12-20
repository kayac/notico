// ここでは初期化のみ行う
// 初期化後は ../lambda で管理している
resource "aws_lambda_function" "lambda" {
  function_name = "notico"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = data.archive_file.nullzip.output_path
  role          = aws_iam_role.lambda.arn

  lifecycle {
    ignore_changes = all
  }
}

resource "aws_lambda_function_url" "lambda" {
  function_name      = aws_lambda_function.lambda.function_name
  authorization_type = "NONE"
}

data "archive_file" "nullzip" {
  type = "zip"
  source {
    content  = "null"
    filename = "bootstrap"

  }
  output_path = "${path.module}/null.zip"

  depends_on = [
    terraform_data.null,
  ]
}

resource "terraform_data" "null" {}
