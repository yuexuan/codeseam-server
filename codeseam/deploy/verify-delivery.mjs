// Real HTTP checks against the running Go service and installed client proxy.
import fs from 'node:fs/promises';
import crypto from 'node:crypto';
const here=new URL('./',import.meta.url), root=new URL('../../../',here);
const output=new URL('docs/tasks/evidence/2026-10-26-P5/',root);
const fixture=JSON.parse(await fs.readFile(new URL('.local/p2-fixture.json',here),'utf8'));
const creds=JSON.parse(await fs.readFile(new URL('.local/test-credentials.json',here),'utf8'));
const c=fixture.clients[0], base='http://127.0.0.1:18081';
const evidence={captured_at:new Date().toISOString(),checks:[]};
function check(name,ok,details={}){evidence.checks.push({name,pass:!!ok,...details});if(!ok)throw Error(name);}
async function api(method,path,body,token,allowed=false){const r=await fetch(base+'/api/v1'+path,{method,headers:{'content-type':'application/json',...(token?{authorization:`Bearer ${token}`}:{})},body:body?JSON.stringify(body):undefined});const v=await r.json();if(allowed)return {http:r.status,data:v.data};if(!r.ok||v.code!==0)throw Error(`HTTP ${r.status} ${path}`);return v.data;}
try{
 const r=await fetch(base+'/portal/');const html=await r.text();check('Go serves production Portal',r.status===200&&html.includes('/portal/assets/')&&!html.includes('/src/'));
 for(const asset of [...html.matchAll(/(?:src|href)="(\/portal\/assets\/[^\"]+)"/g)].map(m=>m[1]))check('production asset '+asset,(await fetch(base+asset)).status===200);
 const setup=await fetch(base+'/downloads/codeseam-windows.exe');const data=Buffer.from(await setup.arrayBuffer());const hash=crypto.createHash('sha256').update(data).digest('hex');
 const file=await fs.readFile(new URL('dist/CodeSeam-windows-x64-setup.exe',here));check('Portal download matches installer',setup.status===200&&hash===crypto.createHash('sha256').update(file).digest('hex'),{sha256:hash,bytes:data.length});
 check('unauthenticated service remains denied',(await api('GET','/codeseam/service/current',null,null,true)).http===401);
 const user=await api('POST','/auth/login',{email:c.email,password:c.password});
 const current=await api('GET','/codeseam/service/current',null,user.access_token);
 const usage=await api('GET','/usage/stats?start_date=2026-10-26&end_date=2026-10-27',null,user.access_token);
 const native=JSON.parse(await fs.readFile(new URL('current-usage.json',output),'utf8')).value;
 check('desktop and Portal API usage agree',native.total_requests===usage.total_requests&&native.total_tokens===usage.total_tokens&&native.total_actual_cost===usage.total_actual_cost,{date_range:['2026-10-26','2026-10-27'],desktop:native,portal_api:usage});
 check('same active service and expiry',current.services.length===1&&current.services[0].id===c.service_id&&current.services[0].state==='ACTIVE',{services:current.services});
 const admin=await api('POST','/auth/login',{email:creds.admin_email,password:creds.admin_password});
 await api('POST',`/admin/subscriptions/${c.subscription_id}/revoke`,{},admin.access_token);
 try{
  const revoked=await api('GET','/codeseam/service/current',null,user.access_token);
  const activation=await api('POST','/codeseam/client/activate',{service_id:c.service_id},user.access_token,true);
  const blocked=await fetch('http://127.0.0.1:18899/claude/v1/messages',{method:'POST',headers:{'content-type':'application/json','x-api-key':'codeseam-local-managed','anthropic-version':'2023-06-01'},body:JSON.stringify({model:'claude-sonnet-4-5',max_tokens:16,messages:[{role:'user',content:'This revoked request must not reach an upstream'}]})});
  check('revoked service denies activation and installed proxy request',activation.http===403&&blocked.status>=400&&revoked.services[0]?.state!=='ACTIVE',{activation_http:activation.http,gateway_http:blocked.status,service_state:revoked.services[0]?.state});
 }finally{await api('POST',`/admin/subscriptions/${c.subscription_id}/restore`,{},admin.access_token);}
 const restored=await api('GET','/codeseam/service/current',null,user.access_token);check('service restored',restored.services[0]?.state==='ACTIVE');
 await api('POST','/auth/logout',{refresh_token:user.refresh_token});
 await api('POST','/auth/logout',{refresh_token:admin.refresh_token});
}finally{await fs.mkdir(output,{recursive:true});await fs.writeFile(new URL('delivery.json',output),JSON.stringify(evidence,null,2));console.log(JSON.stringify(evidence,null,2));}
