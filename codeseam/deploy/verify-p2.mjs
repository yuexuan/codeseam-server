// Real native resources + DeepSeek requests through two authenticated loopback
// CONNECT fixtures. This does NOT certify dedicated public IPs or P3 protection.
import fs from 'node:fs/promises';
import path from 'node:path';
import http from 'node:http';
import net from 'node:net';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const here=path.dirname(fileURLToPath(import.meta.url));
const local=path.join(here,'.local');
const file=path.join(local,'p2-fixture.json');
const base='http://127.0.0.1:18081/api/v1';
const secret=JSON.parse(await fs.readFile(path.join(local,'test-credentials.json'),'utf8'));
async function call(method,route,body,token,allow=false){
 const response=await fetch(base+route,{method,headers:{'content-type':'application/json',...(token?{authorization:`Bearer ${token}`}:{})},body:body===undefined?undefined:JSON.stringify(body)});
 const value=await response.json();
 if(!allow && (!response.ok||value.code!==0))throw Error(`${method} ${route} HTTP ${response.status} (body withheld)`);
 return allow?{status:response.status,value}:value.data;
}
const admin=await call('POST','/auth/login',{email:secret.admin_email,password:secret.admin_password});
const at=admin.access_token;
// DPAPI plaintext exists only in this process; never print/persist it or copy
// the redacted admin DTO as a usable upstream credential.
const upstreamPath=path.join(local,'upstream.dpapi').replaceAll("'","''");
const upstream=JSON.parse(execFileSync('pwsh',['-NoProfile','-Command',`$b=[Security.Cryptography.ProtectedData]::Unprotect([IO.File]::ReadAllBytes('${upstreamPath}'),$null,[Security.Cryptography.DataProtectionScope]::CurrentUser);[Console]::Write([Text.Encoding]::UTF8.GetString($b))`],{encoding:'utf8',windowsHide:true}));
const credentials={base_url:upstream.base_url,api_key:upstream.api_key,model_mapping:{'*':upstream.model}};
let fixture;
try{fixture=JSON.parse(await fs.readFile(file,'utf8'));}catch(e){if(e.code!=='ENOENT')throw e;}
if(!fixture){
 fixture={kind:'loopback CONNECT fixtures; same DeepSeek credential, no dedicated public IP proof',clients:[]};
 const source=await call('GET','/admin/accounts/1',undefined,at);
 if(source.credentials?.base_url!=='https://api.deepseek.com/anthropic')throw Error('Expected P1 DeepSeek fixture');
 for(const [i,label] of ['A','B'].entries()){
  const c={label,email:`p2-${label.toLowerCase()}@codeseam.test`,password:crypto.randomBytes(24).toString('hex'),proxyPassword:crypto.randomBytes(24).toString('hex'),port:18991+i};
  const user=await call('POST','/admin/users',{email:c.email,password:c.password,role:'user',notes:'P2 isolated local QA fixture'},at);c.user_id=user.id;
  const group=await call('POST','/admin/groups',{name:`CodeSeam P2 ${label} local QA`,platform:'anthropic',subscription_type:'subscription',is_exclusive:true,rate_multiplier:1},at);c.group_id=group.id;
  await call('PUT',`/admin/users/${c.user_id}`,{allowed_groups:[group.id]},at);
  const proxy=await call('POST','/admin/proxies',{name:`P2 ${label} LOOPBACK TEST ONLY`,protocol:'http',host:'127.0.0.1',port:c.port,username:`p2-${label}`,password:c.proxyPassword,fallback_mode:'none'},at);c.proxy_id=proxy.id;
  const account=await call('POST','/admin/accounts',{name:`P2 ${label} DeepSeek protocol QA (same upstream test credential)`,platform:'anthropic',type:'apikey',credentials,group_ids:[group.id],proxy_id:proxy.id,concurrency:2,priority:1},at);c.account_id=account.id;
  const sub=await call('POST','/admin/subscriptions/assign',{user_id:user.id,group_id:group.id,validity_days:30,notes:'P2 local test only'},at);c.subscription_id=sub.id;
  const auth=await call('POST','/auth/login',{email:c.email,password:c.password});
  const key=await call('POST','/keys',{name:`CodeSeam Dedicated P2 ${label}`,group_id:group.id,quota:1},auth.access_token);c.service_id=key.id;
  await call('PUT',`/admin/accounts/${account.id}`,{extra:{...account.extra,codeseam_service_id:String(key.id)}},at);
  c.binding={user_id:user.id,service_id:key.id,account_id:account.id,proxy_id:proxy.id,expected_egress:`192.0.2.${i+1}`,revision:'local-fixture-1'};
  await call('PUT',`/admin/groups/${group.id}`,{description:'codeseam:v1:'+JSON.stringify(c.binding)},at);
  fixture.clients.push(c);
  await fs.writeFile(file,JSON.stringify(fixture,null,2));
 }
}
if(fixture.clients.length!==2)throw Error('Incomplete fixture: inspect native resources before resuming');
for(const c of fixture.clients){await call('PUT',`/admin/accounts/${c.account_id}`,{credentials},at);}
if(!fixture.replacement){
 const c={label:'A-replacement',port:18993,proxyPassword:crypto.randomBytes(24).toString('hex')};
 const p=await call('POST','/admin/proxies',{name:'P2 A replacement LOOPBACK TEST ONLY',protocol:'http',host:'127.0.0.1',port:c.port,username:`p2-${c.label}`,password:c.proxyPassword,fallback_mode:'none'},at);
 c.proxy_id=p.id;fixture.replacement=c;await fs.writeFile(file,JSON.stringify(fixture,null,2));
}
function relay(c){
 const state={count:0,targets:[],sockets:new Set()};
 const server=http.createServer((req,res)=>{res.writeHead(405);res.end();});
 server.on('connection',s=>{state.sockets.add(s);s.on('close',()=>state.sockets.delete(s));});
 server.on('connect',(req,socket,head)=>{
  if(req.headers['proxy-authorization']!==`Basic ${Buffer.from(`p2-${c.label}:${c.proxyPassword}`).toString('base64')}`){socket.end('HTTP/1.1 407 Proxy Authentication Required\r\n\r\n');return;}
  if(req.url!=='api.deepseek.com:443'){socket.end('HTTP/1.1 403 Forbidden\r\n\r\n');return;}
  const upstream=net.connect(443,'api.deepseek.com');state.sockets.add(upstream);upstream.on('close',()=>state.sockets.delete(upstream));
  upstream.on('connect',()=>{state.count++;state.targets.push(req.url);socket.write('HTTP/1.1 200 Connection Established\r\n\r\n');if(head.length)upstream.write(head);socket.pipe(upstream);upstream.pipe(socket);});
  upstream.on('error',()=>socket.destroy());socket.on('error',()=>upstream.destroy());socket.on('close',()=>upstream.destroy());
 });
 state.start=()=>new Promise((resolve,reject)=>{server.once('error',reject);server.listen(c.port,'127.0.0.1',resolve);});
 state.stop=()=>new Promise(resolve=>{for(const s of state.sockets)s.destroy();server.close(resolve);});
 return state;
}
const relays=fixture.clients.map(relay);
await Promise.all(relays.map(r=>r.start()));
if(process.argv.includes('--serve')){console.log('P2 loopback CONNECT fixtures ready on 18991/18992; no public IP claim');await new Promise(()=>{});}
const report={time:new Date().toISOString(),scope:fixture.kind,expected_egress:'192.0.2.1/2 are documentation-only placeholders; unverified',checks:[],real_proxy_acceptance:'NON_PASS: no real dedicated proxy/public IP or client process route supplied'};
function assert(ok,name,details={}){report.checks.push({name,pass:!!ok,...details});if(!ok)throw Error(`Check failed: ${name}`);}
const clients=[];
for(const c of fixture.clients){
 const auth=await call('POST','/auth/login',{email:c.email,password:c.password});
 const activation=await call('POST','/codeseam/client/activate',{service_id:c.service_id},auth.access_token);
 const network=await call('POST','/codeseam/client/network-config',{service_id:c.service_id},auth.access_token);
 assert(network.revision===activation.network_revision&&network.proxy.port===c.port&&network.proxy.password===c.proxyPassword,`network config ${c.label}`);
 clients.push({...c,token:auth.access_token,key:activation.api_key});
}
async function request(c){
 const response=await fetch('http://127.0.0.1:18081/v1/messages',{method:'POST',headers:{'content-type':'application/json','x-api-key':c.key,'anthropic-version':'2023-06-01'},body:JSON.stringify({model:'claude-sonnet-4-5',max_tokens:48,stream:true,thinking:{type:'disabled'},messages:[{role:'user',content:`Reply exactly P2_${c.label}_ROUTE_OK`}]})});
 const raw=await response.text();
 const events=raw.split('\n').filter(l=>l.startsWith('data: ')).map(l=>{try{return JSON.parse(l.slice(6));}catch{return {};}});
 return {status:response.status,complete:events.some(e=>e.type==='message_stop'),text:events.filter(e=>e.delta?.type==='text_delta').map(e=>e.delta.text).join(''),request_id:response.headers.get('x-request-id')};
}
const [a,b]=clients;
try{
 for(const c of clients){
  const before=relays.map(r=>r.count);const result=await request(c);const idx=c.label==='A'?0:1;
  assert(result.status===200&&result.complete&&result.text===`P2_${c.label}_ROUTE_OK`&&relays[idx].count>0&&relays[1-idx].count===before[1-idx],`real DeepSeek SSE ${c.label}`,result);
 }
 const foreign=await call('POST','/codeseam/client/network-config',{service_id:b.service_id,user_id:b.user_id},a.token,true);
 assert(foreign.status===404,'foreign config denied');
 const native=async(route,body)=>call('PUT',route,body,at);
 await native(`/admin/groups/${a.group_id}`,{fallback_group_id:b.group_id});
 try{const r=await request(a);assert(r.status===403,'group fallback refused',r);}finally{await native(`/admin/groups/${a.group_id}`,{fallback_group_id:0});}
 await native(`/admin/accounts/${a.account_id}`,{group_ids:[a.group_id,b.group_id]});
 try{const r=await request(a);assert(r.status===403,'shared account refused',r);}finally{await native(`/admin/accounts/${a.account_id}`,{group_ids:[a.group_id]});}
 await native(`/admin/proxies/${a.proxy_id}`,{fallback_mode:'direct'});
 try{const r=await request(a);assert(r.status===403,'direct proxy fallback refused',r);}finally{await native(`/admin/proxies/${a.proxy_id}`,{fallback_mode:'none'});}
 await native(`/admin/accounts/${a.account_id}`,{status:'inactive'});
 try{const r=await request(a);assert(r.status===403,'account disabled',r);}finally{await native(`/admin/accounts/${a.account_id}`,{status:'active'});}
 await call('PUT',`/keys/${a.service_id}`,{status:'inactive'},a.token);
 try{const r=await request(a);assert(r.status>=400,'key disabled',r);}finally{await call('PUT',`/keys/${a.service_id}`,{status:'active'},a.token);}
 await call('PUT',`/keys/${a.service_id}`,{expires_at:new Date(Date.now()-60000).toISOString()},a.token);
 try{const r=await request(a);assert(r.status>=400,'service key expired',r);const v=await call('POST','/codeseam/client/network-config',{service_id:a.service_id},a.token,true);assert(v.status===403,'expired config denied');}finally{await call('PUT',`/keys/${a.service_id}`,{expires_at:''},a.token);}
 const previous=await call('GET',`/admin/subscriptions/${a.subscription_id}`,undefined,at);
 const extended=await call('POST',`/admin/subscriptions/${a.subscription_id}/extend`,{days:1},at);
 assert(Date.parse(extended.expires_at)>Date.parse(previous.expires_at),'native subscription renewal');
 await call('POST',`/admin/subscriptions/${a.subscription_id}/revoke`,{},at);
 try{const r=await request(a);assert(r.status>=400,'subscription revoked',r);}finally{await call('POST',`/admin/subscriptions/${a.subscription_id}/restore`,{},at);}
 const replacement=relay(fixture.replacement);await replacement.start();
 const oldConfig=await call('POST','/codeseam/client/network-config',{service_id:a.service_id},a.token);
 try{
  await native(`/admin/accounts/${a.account_id}`,{proxy_id:fixture.replacement.proxy_id});
  assert((await request(a)).status===403,'partial replacement fails closed');
  const binding={...a.binding,proxy_id:fixture.replacement.proxy_id,revision:'replacement-test'};
  await native(`/admin/groups/${a.group_id}`,{description:'codeseam:v1:'+JSON.stringify(binding)});
  const next=await call('POST','/codeseam/client/network-config',{service_id:a.service_id},a.token);
  assert(next.revision!==oldConfig.revision&&next.proxy.port===18993,'replacement config revision and endpoint');
  const result=await request(a);assert(result.status===200&&result.complete&&replacement.count>0,'replacement proxy real SSE',result);
 }finally{
  await native(`/admin/accounts/${a.account_id}`,{proxy_id:a.proxy_id});
  await native(`/admin/groups/${a.group_id}`,{description:'codeseam:v1:'+JSON.stringify(a.binding)});
  await replacement.stop();
 }
 const otherBefore=relays[1].count;
 await relays[0].stop();
 try{const r=await request(a);assert(r.status>=400&&relays[1].count===otherBefore,'proxy disconnected without other customer fallback',r);}finally{await relays[0].start();}
 // Native scheduler may temporarily pause after a transport failure.
 await call('POST',`/admin/accounts/${a.account_id}/clear-error`,{},at);
 await native(`/admin/accounts/${a.account_id}`,{schedulable:true,status:'active'});
 const recovered=await request(a);assert(recovered.status===200&&recovered.complete,'renewed/restored service real SSE',recovered);
 for(const c of clients){const stats=await call('GET','/usage/stats',undefined,c.token);report.checks.push({name:`native usage ${c.label}`,stats});}
}finally{
 await Promise.all(relays.map(r=>r.stop()));
 report.connects=relays.map((r,i)=>({client:clients[i].label,count:r.count,targets:[...new Set(r.targets)]}));
 const output=path.resolve(here,'../../../docs/tasks/evidence/2026-10-26-P2/verification.json');
 await fs.mkdir(path.dirname(output),{recursive:true});await fs.writeFile(output,JSON.stringify(report,null,2));
 console.log(JSON.stringify(report,null,2));
}
