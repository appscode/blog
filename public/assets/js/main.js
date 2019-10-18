document.addEventListener("DOMContentLoaded", () => {
  // Get all "navbar-burger" elements
  const $navbarBurgers = Array.prototype.slice.call(
    document.querySelectorAll(".navbar-burger"),
    0
  );
  // Check if there are any navbar burgers
  if ($navbarBurgers.length > 0) {
    // Add a click event on each of them
    $navbarBurgers.forEach(el => {
      el.addEventListener("click", () => {
        // Get the target from the "data-target" attribute
        const target = el.dataset.target;
        const $target = document.getElementById(target);

        // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
        el.classList.toggle("is-active");
        $target.classList.toggle("is-active");
      });
    });
  }
});

// menu sticky
//Not a ton of code, but hard to
const nav = document.querySelector("#header");
let topOfNav = nav.offsetTop + 1;
function fixNav() {
  if (window.scrollY >= topOfNav) {
    document.body.classList.add("fixed-nav");
  } else {
    document.body.classList.remove("fixed-nav");
    document.body.style.paddingTop = 0;
  }
}
window.addEventListener("scroll", fixNav);

// scroll to top
var basicScrollTop = function() {
  // The button
  var btnTop = document.querySelector("#goTop");
  // Reveal the button
  var btnReveal = function() {
    if (window.scrollY >= 300) {
      btnTop.classList.add("is-visible");
    } else {
      btnTop.classList.remove("is-visible");
    }
  };
  // Smooth scroll top
  var TopscrollTo = function() {
    if (window.scrollY != 0) {
      setTimeout(function() {
        window.scrollTo(0, window.scrollY - 30);
        TopscrollTo();
      }, 5);
    }
  };
  // Listeners
  window.addEventListener("scroll", btnReveal);
  btnTop.addEventListener("click", TopscrollTo);
};
basicScrollTop();

// blog page hero slider
var sBtn = document.getElementById("sBtn");
if(sBtn){
  sBtn.addEventListener("click", function(event) {
    event.preventDefault();
  
    const sliderItems = document.querySelectorAll(".single-blog");
    const arr = Array.from(sliderItems);
  
    let indexOfShow = 0;
    arr.forEach((sliderItem, idx) => {
      if (sliderItem.classList.contains("show")) {
        indexOfShow = idx;
        sliderItem.classList.remove("show");
      }
    });
  
    let newIndex = (indexOfShow + 1) % arr.length;
    arr[newIndex].classList.add("show");
  });
}


//clipboard js
var copyBtn = document.querySelector('.copy-here')
if(copyBtn){
  new ClipboardJS(copyBtn);
  copyBtn.addEventListener('click', function(){
    copyBtn.setAttribute("title", "copied!")
  })
}