// 値はAWS Console上で手動設定しています
resource "aws_secretsmanager_secret" "secrets" {
  name = "/notico/prod"
}
