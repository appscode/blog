const keywords = new Set();
let searchKeyword = "";
let viewType = "grid-view";

const searchElement = document.getElementById("search");
const gridView = document.getElementById("grid-view");
const listView = document.getElementById("list-view");
const gridViewBtn = document.getElementById("grid-btn-view");
const listViewBtn = document.getElementById("list-btn-view");
const productElement = document.getElementById("products");
const categoriesElement = document.getElementById("categories");


gridViewBtn?.addEventListener("click", (event) => {
  gridViewBtn.classList.add("is-active");
  gridView.classList.remove("is-hidden");
  listViewBtn.classList.remove("is-active");
  listView.classList.add("is-hidden");
  viewType = "grid-view";
  filterList();
})

listViewBtn?.addEventListener("click", (event) => {
  listViewBtn.classList.add("is-active");
  listView.classList.remove("is-hidden");
  gridViewBtn.classList.remove("is-active");
  gridView.classList.add("is-hidden");
  viewType = "list-view";
  filterList();
})

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
  let str = searchElement.value;
  searchKeyword = str.toLowerCase();
  filterList();
})

//Filter based on the tags and search keyword
const filterList = () => {
  const cards = document.getElementById(viewType);
  const cardList = cards.querySelectorAll(".each-blog");
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
      card.style.display = viewType === "grid-view" ? "block" : "flex";
    }
  })
}

//Check if tags & search keyword contains in cards tags, auther and heading
const isTagAvailable = (tags, author, heading) => {
  if (keywords.size === 0 && searchKeyword.length < 3) return true;
  let flag = false;
  if (keywords.size === 0) {
    flag |= tags.includes(searchKeyword);
    flag |= author.includes(searchKeyword);
    flag |= heading.includes(searchKeyword);
  } else if (searchKeyword.length < 3) {
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