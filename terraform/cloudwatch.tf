resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/notico"
  retention_in_days = 0
}
