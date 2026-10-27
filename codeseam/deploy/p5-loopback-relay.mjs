// Actual TCP/HTTP forwarding for installed-client QA. No synthetic egress/SSE.
// Shared machine egress is NOT proof of a dedicated public proxy or P3 coverage.
import fs from 'node:fs/promises';
import http from 'node:http';
import net from 'node:net';
import {execFileSync} from 'node:child_process';
const local = new URL('./.local/', import.meta.url);
const fixture = JSON.parse(await fs.readFile(new URL('p2-fixture.json',local),'utf8'));
const secret = JSON.parse(await fs.readFile(new URL('test-credentials.json',local),'utf8'));
const c = fixture.clients[0];
async function api(method,path,body,token){
 const r=await fetch('http://127.0.0.1:18081/api/v1'+path,{method,headers:{'content-type':'application/json',authorization:`Bearer ${token||''}`},body:body?JSON.stringify(body):undefined});
 const v=await r.json(); if(!r.ok||v.code!==0)throw Error(`QA resource API HTTP ${r.status}`);return v.data;
}
const auth=await api('POST','/auth/login',{email:secret.admin_email,password:secret.admin_password});
const backup=new URL('p5-group-backup.json',local);
if(process.argv.includes('--restore')){
 const old=JSON.parse(await fs.readFile(backup,'utf8'));
 await api('PUT',`/admin/groups/${c.group_id}`,{description:old.description},auth.access_token);
 console.log('Original QA binding restored');process.exit(0);
}
const group=await api('GET',`/admin/groups/${c.group_id}`,null,auth.access_token);
try{await fs.access(backup);}catch{await fs.writeFile(backup,JSON.stringify({description:group.description}));}
const binding=JSON.parse(group.description.slice('codeseam:v1:'.length));
const proxy=await api('GET',`/admin/proxies/${binding.proxy_id}`,null,auth.access_token);
const port=proxy.port;
const sockets=new Set();let count=0;
const authorized=req=>req.headers['proxy-authorization']===`Basic ${Buffer.from(`${proxy.username}:${c.proxyPassword}`).toString('base64')}`;
const server=http.createServer((req,res)=>{
 if(!authorized(req)){res.writeHead(407);res.end();return;}
 let url;try{url=new URL(req.url);}catch{res.writeHead(400);res.end();return;}
 if(url.origin!=='http://127.0.0.1:18081'){res.writeHead(403);res.end();return;}
 count++;
 const headers={...req.headers};delete headers['proxy-authorization'];
 const upstream=http.request(url,{method:req.method,headers},r=>{res.writeHead(r.statusCode,r.headers);r.pipe(res);});
 upstream.on('error',()=>{if(!res.headersSent)res.writeHead(502);res.end();});req.pipe(upstream);
});
server.on('connection',s=>{sockets.add(s);s.on('close',()=>sockets.delete(s));});
server.on('connect',(req,socket,head)=>{
 if(!authorized(req)){socket.end('HTTP/1.1 407 Proxy Authentication Required\r\n\r\n');return;}
 if(!['api.deepseek.com:443','api.ipify.org:443','api6.ipify.org:443'].includes(req.url)){socket.end('HTTP/1.1 403 Forbidden\r\n\r\n');return;}
 const [host,p]=req.url.split(':');const upstream=net.connect(Number(p),host);sockets.add(upstream);
 upstream.on('connect',()=>{count++;socket.write('HTTP/1.1 200 Connection Established\r\n\r\n');if(head.length)upstream.write(head);socket.pipe(upstream);upstream.pipe(socket);});
 upstream.on('error',()=>socket.destroy());socket.on('error',()=>upstream.destroy());socket.on('close',()=>upstream.destroy());upstream.on('close',()=>sockets.delete(upstream));
});
await new Promise((ok,fail)=>{server.once('error',fail);server.listen(port,'127.0.0.1',ok);});
// Async curl: the relay must stay on the event loop while the real TLS request runs.
const {execFile}=await import('node:child_process');
const proxyUrl=`http://${encodeURIComponent(proxy.username)}:${encodeURIComponent(c.proxyPassword)}@127.0.0.1:${port}`;
const observed=await new Promise((ok,fail)=>execFile('curl.exe',['--silent','--show-error','--max-time','20','--proxy',proxyUrl,'https://api.ipify.org'],{windowsHide:true},(err,out)=>err?fail(Error('Real egress probe failed (credentials withheld)')):ok(out.trim())));
if(net.isIP(observed)!==4)throw Error('Probe did not return a real IPv4 address');
binding.expected_egress=observed;binding.revision='p5-shared-loopback-real-probe';
await api('PUT',`/admin/groups/${c.group_id}`,{description:'codeseam:v1:'+JSON.stringify(binding)},auth.access_token);
await fs.writeFile(new URL('p5-relay.json',local),JSON.stringify({pid:process.pid,port,scope:'authenticated loopback TCP relay; real shared host egress; NOT dedicated public egress'}));
console.log(`P5 real forwarding relay ready on ${port}; shared egress observed; no dedicated IP claim`);
process.on('SIGINT',()=>{console.log(`Relay forwarded ${count} connections`);for(const s of sockets)s.destroy();server.close(()=>process.exit(0));});
await new Promise(()=>{});
