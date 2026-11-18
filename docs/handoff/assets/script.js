(function(){
  var root=document.documentElement, osKey="handoff-os", themeKey="handoff-theme";
  function store(k,v){try{localStorage.setItem(k,v);}catch(e){}}
  function read(k){try{return localStorage.getItem(k);}catch(e){return null;}}
  function guess(){
    var p=((navigator.userAgentData&&navigator.userAgentData.platform)||navigator.platform||navigator.userAgent||"").toLowerCase();
    if(p.indexOf("win")>-1) return "windows";
    if(p.indexOf("mac")>-1) return "mac";
    return "linux";
  }
  function setOs(os){
    root.setAttribute("data-os",os);
    document.querySelectorAll(".os-bar button").forEach(function(b){b.setAttribute("aria-pressed",String(b.dataset.os===os));});
    store(osKey,os);
  }
  setOs(read(osKey)||guess());
  document.querySelectorAll(".os-bar button").forEach(function(b){b.addEventListener("click",function(){setOs(b.dataset.os);});});
  var saved=read(themeKey);
  if(saved==="dark"||saved==="light") root.setAttribute("data-theme",saved);
  var toggle=document.querySelector("[data-theme-toggle]");
  if(toggle){
    toggle.addEventListener("click",function(){
      var cur=root.getAttribute("data-theme")||(window.matchMedia&&window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light");
      var next=cur==="dark"?"light":"dark";
      root.setAttribute("data-theme",next);
      store(themeKey,next);
    });
  }
  function copyText(t){
    if(navigator.clipboard&&window.isSecureContext){return navigator.clipboard.writeText(t);}
    return new Promise(function(ok,fail){var a=document.createElement("textarea");a.value=t;a.setAttribute("readonly","");a.style.position="fixed";a.style.opacity="0";document.body.appendChild(a);a.select();var r=false;try{r=document.execCommand("copy");}catch(e){}document.body.removeChild(a);r?ok():fail();});
  }
  document.querySelectorAll(".codeblock .copy").forEach(function(btn){
    btn.addEventListener("click",function(){
      var t=btn.closest(".codeblock").querySelector("pre code").innerText;
      copyText(t).then(function(){btn.textContent="Copied";btn.classList.add("done");setTimeout(function(){btn.textContent="Copy";btn.classList.remove("done");},1500);}).catch(function(){btn.textContent="Select and copy";});
    });
  });
  document.querySelectorAll("pre").forEach(function(pre){
    if(pre.closest(".codeblock")) return;
    var btn=document.createElement("button"); btn.className="copy"; btn.type="button"; btn.textContent="Copy";
    btn.addEventListener("click",function(){
      var c=pre.querySelector("code")||pre;
      copyText(c.innerText).then(function(){btn.textContent="Copied";setTimeout(function(){btn.textContent="Copy";},1500);}).catch(function(){btn.textContent="Select and copy";});
    });
    pre.appendChild(btn);
  });
})();
