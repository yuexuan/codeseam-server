# Local development fixture using ONLY native auth/admin/key APIs.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$UpstreamFile,[switch]$AcceptLocalCompliance)
$ErrorActionPreference='Stop'
$runtime=Join-Path $PSScriptRoot '.local'
$secret=Get-Content (Join-Path $runtime 'test-credentials.json') -Raw|ConvertFrom-Json
if ([IO.Path]::GetExtension($UpstreamFile) -eq '.dpapi') {
    $bytes=[Security.Cryptography.ProtectedData]::Unprotect([IO.File]::ReadAllBytes($UpstreamFile),$null,[Security.Cryptography.DataProtectionScope]::CurrentUser)
    $upstream=[Text.Encoding]::UTF8.GetString($bytes)|ConvertFrom-Json
} else {$upstream=Get-Content -LiteralPath $UpstreamFile -Raw|ConvertFrom-Json}
$base='http://127.0.0.1:18081/api/v1'
function Api($method,$route,$body,$headers) {
    $p=@{Method=$method;Uri="$base$route";ContentType='application/json'}
    if($body){$p.Body=$body|ConvertTo-Json -Depth 80}
    if($headers){$p.Headers=$headers}
    try { $r=Invoke-RestMethod @p } catch {throw "Native API $method $route failed (response withheld to protect credentials)"}
    if($r.code -ne 0){throw "Native API $method $route rejected"};return $r.data
}
$admin=Api Post '/auth/login' @{email=$secret.admin_email;password=$secret.admin_password} $null
$ah=@{Authorization="Bearer $($admin.access_token)"}
if($AcceptLocalCompliance){$null=Api Post '/admin/compliance/accept' @{phrase='我已阅读、理解并同意 Sub2API 部署与运营合规承诺';language='zh'} $ah}
$settings=Api Get '/admin/settings' $null $ah
$settings.registration_enabled=$true
$settings.email_verify_enabled=$false
$null=Api Put '/admin/settings' $settings $ah
try {$null=Api Post '/auth/register' @{email=$secret.user_email;password=$secret.user_password} $null} catch {
    # A repeated run may find an existing user. Native login below must succeed.
}
$user=Api Post '/auth/login' @{email=$secret.user_email;password=$secret.user_password} $null
$uh=@{Authorization="Bearer $($user.access_token)"}
$current=Api Get '/codeseam/service/current' $null $uh
if($current.services.Count -ne 0){throw 'Test user already has resources; inspect instead of provisioning duplicates'}
$group=Api Post '/admin/groups' @{name='CodeSeam P1 DeepSeek temporary';platform='anthropic';subscription_type='standard';rate_multiplier=1;is_exclusive=$true;claude_code_only=$false} $ah
$uid=$user.user.id
$null=Api Put "/admin/users/$uid" @{allowed_groups=@($group.id)} $ah
$null=Api Post "/admin/users/$uid/balance" @{balance=5;operation='set';notes='P1 local development fixture only'} $ah
$account=Api Post '/admin/accounts' @{name='P1 DeepSeek official Anthropic compatibility';platform='anthropic';type='apikey';credentials=@{base_url=$upstream.base_url;api_key=$upstream.api_key;model_mapping=@{'*'=$upstream.model}};group_ids=@($group.id);concurrency=2;priority=1;rate_multiplier=1} $ah
$key=Api Post '/keys' @{name='CodeSeam P1 managed';group_id=$group.id;quota=1;expires_in_days=7} $uh
[ordered]@{user_id=$uid;group_id=$group.id;account_id=$account.id;service_id=$key.id;upstream=$upstream.base_url;model=$upstream.model} |ConvertTo-Json |Set-Content (Join-Path $runtime 'resource-ids.json')
Write-Host "Native registration and resources provisioned: user=$uid group=$($group.id) account=$($account.id) service=$($key.id). No keys printed."
