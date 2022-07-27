

document.addEventListener("DOMContentLoaded", () => {
  // AOS initialization
  AOS.init({
    once: true,
  });

  // navbar for mobile device
  let navbar = document.querySelector(".navbar-burger");
  navbar?.addEventListener("click", function () {
    const hasActiveClass = navbar.classList.contains("is-active");
    let dropdown = document.querySelector(".navbar-right");
    navbar.classList.toggle("is-active");
    dropdown.style.opacity = 1 - dropdown.style.opacity;
    dropdown.style.visibility = hasActiveClass ? "hidden" : "visible";
  });
});

// scroll to top
var basicScrollTop = function () {
  // The button
  var btnTop = document.querySelector("#goTop");
  // Reveal the button
  var btnReveal = function () {
    if (window.scrollY >= 300) {
      btnTop.classList.add("is-visible");
    } else {
      btnTop.classList.remove("is-visible");
    }
  };
  // Smooth scroll top
  var TopscrollTo = function () {
    if (window.scrollY != 0) {
      setTimeout(function () {
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
var carouselBtn = document.getElementsByClassName("carousel-button");
if (carouselBtn) {
  Array.from(carouselBtn).forEach((carouselEl, carouselElIdx) => {
    carouselEl.addEventListener("click", function (event) {
      console.log(carouselElIdx);
      event.preventDefault();
      const sliderItems = document.querySelectorAll(".single-blog-carousel");
      const arr = Array.from(sliderItems);

      let indexOfShow = carouselElIdx;
      arr.forEach((sliderItem, idx) => {
        if (sliderItem.classList.contains("show")) {
          indexOfShow = idx;
          sliderItem.classList.remove("show");
        }
      });

      let newIndex = (indexOfShow + 1) % arr.length;
      arr[newIndex].classList.add("show");
    });
  });
}

// code download and copy function //
var codeHeading = document.querySelectorAll(".code-block-heading");
Array.from(codeHeading).forEach((heading) => {
  const pre = heading.nextElementSibling;
  const code = pre.querySelector("code");
  const codeContent = code.textContent;
  let fileType = code.getAttribute("class");
  if (fileType) {
    fileType = fileType.replace("language-", "");
  } else {
    fileType = "txt";
  }
  let fileName = heading
    .querySelector(".code-title > h4")
    .textContent.replace(" ", "_");

  // download js //
  var downloadBtn = heading.querySelector(".download-here");
  if (downloadBtn) {
    downloadBtn.addEventListener("click", function () {
      return download(codeContent, `${fileName}.${fileType}`, "text/plain");
    });
  }

  //clipboard js
  var copyBtn = heading.querySelector(".copy-here");
  if (copyBtn) {
    new ClipboardJS(copyBtn);
    copyBtn.addEventListener("click", function () {
      copyBtn.setAttribute("title", "copied!");
    });
  }
});

// tabs active class add script - setup | install page
const tabItems = document.querySelectorAll(".nav-item .nav-link");
tabItems.forEach((tab) => {
  tab.addEventListener("click", (e) => {
    e.preventDefault();
    const el = e.currentTarget;

    // add .active class to the clicked item, remove .active from others
    document.querySelectorAll(".nav-item .nav-link").forEach((navLink) => {
      navLink === el
        ? navLink.classList.add("active")
        : navLink.classList.remove("active");
    });

    // add .show class to the target tab-pane, remove from others
    const elHref = el.getAttribute("href");
    const tabPaneTarget = document.querySelector(elHref);

    document.querySelectorAll(".tab-pane").forEach((tabPane) => {
      tabPane === tabPaneTarget
        ? tabPane.classList.add("show")
        : tabPane.classList.remove("show");
    });
  });
});




// // blog hero area carousel start
$('.owl-carousel').owlCarousel({
    loop:true,
    dots: false,
    nav: true,
    animateOut: 'fadeOut',
    margin:0,
    infinity: true,
    autoplay: true,
    responsiveClass:true,
    responsive:{
        0:{
            items:1,
          
        },
        600:{
            items:1,
        },
        1000:{
            items:1,
        }
    }
}) 
// // blog hero area carousel end