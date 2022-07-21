const keywords =  new Set();

const kubedb = document.getElementById("kubedb");
const stash = document.getElementById("stash");
const kubevault = document.getElementById("kubevault");
const kubeform = document.getElementById("kubeform");

kubedb.addEventListener("change",(Event)=>{
  let isChecked = Event.target.checked; 
  if(isChecked) keywords.add("kubedb");
  else keywords.delete("kubedb");
  filterList();
})

stash.addEventListener("change",(Event)=>{
  let isChecked = Event.target.checked; 
  if(isChecked) keywords.add("stash");
  else keywords.delete("stash")
  filterList();
})

kubevault.addEventListener("change",(Event)=>{
  let isChecked = Event.target.checked; 
  if(isChecked) keywords.add("kubevault");
  else keywords.delete("kubevault");
  filterList();
})

kubeform.addEventListener("change",(Event)=>{
  let isChecked = Event.target.checked; 
  if(isChecked) keywords.add("kubeform");
  else keywords.delete("kubeform");
  filterList();
})


const filterList = () =>{
  const cards = document.getElementById("card-list");
  const cardList = cards.querySelectorAll(".column");
  cardList.forEach(card => {
    const tags = card.querySelector(".tags").innerText;
    if(!isTagAvailable(tags.toLowerCase())){
      card.style.display = "none";
    }else card.style.display = "block"
  })
}

const isTagAvailable = (tags) =>{
  if(keywords.size === 0) return true;
  let flag = false;
  keywords.forEach(key=>{
    flag |= tags.includes(key);
  })
  return flag;
}