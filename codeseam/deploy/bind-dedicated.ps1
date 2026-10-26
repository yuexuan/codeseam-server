# Thin operator adapter over native admin APIs. Create user/subscription/key,
# account and proxy in Sub2API first. This command writes only the missing relation.
[CmdletBinding()]
param(
 [string]$Server='http://127.0.0.1:18081',
 [string]$CredentialFile=(Join-Path $PSScriptRoot '.local/test-credentials.json'),
 [Parameter(Mandatory)][long]$UserId,
 [Parameter(Mandatory)][long]$ServiceId,
 [Parameter(Mandatory)][long]$GroupId,
 [Parameter(Mandatory)][long]$AccountId,
 [Parameter(Mandatory)][long]$ProxyId,
 [Parameter(Mandatory)][string]$ExpectedEgress,
 [string]$Revision=([Guid]::NewGuid().ToString())
)
$ErrorActionPreference='Stop'
$uri=[uri]$Server
if($uri.Scheme -ne 'https' -and -not ($uri.Scheme -eq 'http' -and $uri.IsLoopback)){throw 'HTTPS required outside loopback'}
$address=$null
if(-not [net.ipaddress]::TryParse($ExpectedEgress,[ref]$address)){throw 'ExpectedEgress must be an operator-verified IP address'}
$credentials=Get-Content -LiteralPath $CredentialFile -Raw|ConvertFrom-Json
$headers=@{}
function Api($method,$route,$body=$null){
 $args=@{Method=$method;Uri="$($Server.TrimEnd('/'))/api/v1$route";Headers=$headers;ContentType='application/json'}
 if($null -ne $body){$args.Body=$body|ConvertTo-Json -Depth 30}
 try{$response=Invoke-RestMethod @args}catch{throw "Native API $method $route failed; response withheld"}
 if($response.code -ne 0){throw "Native API $method $route rejected"}
 return $response.data
}
$login=Api Post '/auth/login' @{email=$credentials.admin_email;password=$credentials.admin_password}
$headers.Authorization="Bearer $($login.access_token)"
$group=Api Get "/admin/groups/$GroupId"
$account=Api Get "/admin/accounts/$AccountId"
$proxy=Api Get "/admin/proxies/$ProxyId"
if($group.platform -ne 'anthropic' -or $group.subscription_type -ne 'subscription' -or -not $group.is_exclusive){throw 'Requires an exclusive native Anthropic subscription group'}
if($group.status -ne 'active'){throw 'Group must be active before binding'}
if($group.description -and -not $group.description.StartsWith('codeseam:')){throw 'Group has an existing description; preserve it before dedicating this group'}
$found=$false
for($page=1;;$page++){
 $keys=Api Get "/admin/users/$UserId/api-keys?page=$page&page_size=100"
 foreach($key in $keys.items){if($key.id -eq $ServiceId -and $key.group_id -eq $GroupId){$found=$true}}
 if($page*100 -ge $keys.total){break}
}
if(-not $found){throw 'Service key must already belong to this user and group'}
if($account.platform -ne 'anthropic' -or $account.parent_account_id){throw 'Requires an independent Anthropic account'}
if($proxy.status -ne 'active'){throw 'Proxy must be active'}
$others=Api Get "/admin/proxies/$ProxyId/accounts"
foreach($other in $others){if($other.id -ne $AccountId){throw 'Proxy is already assigned to another account'}}
$binding=@{user_id=$UserId;service_id=$ServiceId;account_id=$AccountId;proxy_id=$ProxyId;expected_egress=$ExpectedEgress;revision=$Revision}
# Fail closed during a multi-resource operator change. A partial failure leaves
# the native group inactive for inspection rather than rolling into shared routing.
$null=Api Put "/admin/groups/$GroupId" @{status='inactive'}
try{
 $extra=@{}
 if($account.extra){foreach($property in $account.extra.PSObject.Properties){$extra[$property.Name]=$property.Value}}
 $extra.codeseam_service_id="$ServiceId"
 $null=Api Put "/admin/proxies/$ProxyId" @{fallback_mode='none';backup_proxy_id=$null}
 $null=Api Put "/admin/accounts/$AccountId" @{group_ids=@($GroupId);proxy_id=$ProxyId;extra=$extra}
 $null=Api Put "/admin/groups/$GroupId" @{description=('codeseam:v1:'+($binding|ConvertTo-Json -Compress));fallback_group_id=0;fallback_group_id_on_invalid_request=0;model_routing_enabled=$false;status='active'}
}catch{throw 'Binding incomplete; group remains inactive. Inspect native resources before enabling it.'}
Write-Host "Dedicated relation saved: user=$UserId service=$ServiceId group=$GroupId account=$AccountId proxy=$ProxyId. Client must refresh and verify actual egress."
