const keywords =  new Set();
let searchKeyword = "";

const kubedb = document.getElementById("kubedb");
const stash = document.getElementById("stash");
const kubevault = document.getElementById("kubevault");
const kubeform = document.getElementById("kubeform");
const searchElement = document.getElementById("search");

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

searchElement.addEventListener("input",(event)=>{
  let str = searchElement.value;
  searchKeyword = str.toLowerCase();
  filterList();
})

//Filter based on the tags and search keyword
const filterList = () =>{
  const cards = document.getElementById("card-list");
  const cardList = cards.querySelectorAll(".column");
  cardList.forEach(card => {
    const tags = card.querySelector(".tags").innerText.toLowerCase();
    const author = card.querySelector(".author").innerText.toLowerCase();
    const heading = card.querySelector("h2").innerText.toLowerCase();
    if(!isTagAvailable(tags,author,heading)){
      card.style.display = "none";
    }else card.style.display = "block"
  })
}

//Check if tags & search keyword contains in cards tags, auther and heading
const isTagAvailable = (tags,author,heading) =>{
  if(keywords.size === 0 && searchKeyword.length<3) return true;
  let flag = false;
  if(keywords.size === 0){
    flag |= tags.includes(searchKeyword);
    flag |= author.includes(searchKeyword);
    flag |= heading.includes(searchKeyword);
  }else if(searchKeyword.length<3){
    keywords.forEach(key=>{
      flag |= tags.includes(key);
    })
  }else{
    keywords.forEach(key=>{
      let flag1 = tags.includes(key) 
      let flag2 = tags.includes(searchKeyword) || author.includes(searchKeyword) || heading.includes(searchKeyword);
      flag = flag1 & flag2;
    })
  }return flag;
}