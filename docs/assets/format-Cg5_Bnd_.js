const r=t=>{if(!t)return"0 Bytes";const o=1024,e=["Bytes","KB","MB","GB"],a=Math.floor(Math.log(t)/Math.log(o));return`${parseFloat((t/Math.pow(o,a)).toFixed(2))} ${e[a]}`};export{r as f};
