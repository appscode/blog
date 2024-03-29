const keywords = new Set();
let searchKeyword = "";
let viewType = "grid-view";

const searchElement = document.getElementById("search");
const productElement = document.getElementById("products");
const categoriesElement = document.getElementById("categories");
const productBtn = document.getElementById("product-btn");
const categoriesBtn = document.getElementById("categories-btn");
const nodataElement = document.getElementById("nodata-content");

productElement?.addEventListener("change",(event)=>{
  let elementName = event.target.id || "";
  if(elementName){
    let isChecked = event.target.checked || false;
    if (isChecked) keywords.add(elementName.toLowerCase());
    else keywords.delete(elementName.toLowerCase());
    filterList();
  }
})

categoriesElement?.addEventListener("change",(event)=>{
  let elementName = event.target.id || "";
  if(elementName){
    let isChecked = event.target.checked || false;
    if (isChecked) keywords.add(elementName.toLowerCase());
    else keywords.delete(elementName.toLowerCase());
    filterList();
  }
})

searchElement?.addEventListener("input", (event) => {
  searchElement.style.top = 0;
  let str = searchElement.value;
  searchKeyword = str.toLowerCase();
  filterList();
})

//Toggle view of sidebar product list
productBtn?.addEventListener("click",(event)=>{
  const iconElement = productBtn.querySelector("i");
  const className = iconElement.className;
  if(className === "fa fa-angle-up"){
    iconElement.classList.remove("fa-angle-up");
    iconElement.classList.add("fa-angle-down");
    productElement.style.display = "none";

  }else{
    iconElement.classList.remove("fa-angle-down");
    iconElement.classList.add("fa-angle-up")
    productElement.style.display = "block";
  }
})

//Toggle view of sidebar categories list
categoriesBtn?.addEventListener("click",(event)=>{
  const iconElement = categoriesBtn.querySelector("i");
  const className = iconElement.className;
  if(className === "fa fa-angle-up"){
    iconElement.classList.remove("fa-angle-up");
    iconElement.classList.add("fa-angle-down");
    categoriesElement.style.display = "none";

  }else{
    iconElement.classList.remove("fa-angle-down");
    iconElement.classList.add("fa-angle-up")
    categoriesElement.style.display = "block";
  }
})


//Filter based on the tags and search keyword
const filterList = () => {
  window.scroll({
    top: calculateTopVaue(), 
    left: 0, 
    behavior: 'smooth'
  });
  const cards = document.getElementById(viewType);
  const cardList = cards.querySelectorAll(".each-blog");
  let noDataAvailable = true;
  cardList.forEach(card => {
    //get all tags
    const tags = card.querySelector(".tags").innerText.toLowerCase();

    //get all authers name
    let authors = "";
    const authorList = card.querySelectorAll(".author");
    authorList.forEach(author => authors += author.innerText.toLowerCase());

    //get all headings
    const heading = card.querySelector("h2").innerText.toLowerCase();

    if (!isTagAvailable(tags, authors, heading)) {
      card.style.display = "none";
    } else {
      noDataAvailable = false;
      card.style.display = viewType === "grid-view" ? "block" : "flex";
    }
  })
  if(noDataAvailable){
    nodataElement.classList.remove("is-hidden");
  }else{
    nodataElement.classList.add("is-hidden");
  }
}

//Check if tags & search keyword contains in cards tags, auther and heading
const isTagAvailable = (tags, author, heading) => {
  if (keywords.size === 0 && searchKeyword.length <=0) return true;
  let flag = false;
  if (keywords.size === 0) {
    flag |= tags.includes(searchKeyword);
    flag |= author.includes(searchKeyword);
    flag |= heading.includes(searchKeyword);
  } else if (searchKeyword.length <=0) {
    let temFlag = true;
    keywords.forEach(key => {
      temFlag &= tags.includes(key);
    })
    flag = temFlag;
  } else {
    let flag1 = true;
    let flag2 = tags.includes(searchKeyword) || author.includes(searchKeyword) || heading.includes(searchKeyword);
    keywords.forEach(key => {
      flag1 &= tags.includes(key)
    })
    flag = flag1 & flag2;
  }
  return flag;
}




const calculateTopVaue = () =>{
  const heroArea = document.querySelector('.hero-area-blog')
  const recentBlog = document.querySelector('.recent-blog-posts')
  const authorHeroArea = document.querySelector('.author-hero-area')
  let height = (heroArea ? heroArea.offsetHeight : 0) + (recentBlog ? recentBlog.offsetHeight : 0) + (authorHeroArea ? authorHeroArea.offsetHeight : 0);
  return height;
}


//From Mobile View only
if(window.innerWidth<768){
  const pdtElement = productBtn.querySelector("i");
  pdtElement.classList.remove("fa-angle-up");
  pdtElement.classList.add("fa-angle-down");
  productElement.style.display = "none";

  const ctgElement = categoriesBtn.querySelector("i");
  ctgElement.classList.remove("fa-angle-up");
  ctgElement.classList.add("fa-angle-down");
  categoriesElement.style.display = "none";
}