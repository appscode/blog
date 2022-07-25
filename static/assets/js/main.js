document.addEventListener("DOMContentLoaded", () => {
  // AOS initialization
  AOS.init({
    once: false,
  });

  // Get all "navbar-burger" elements
  const $navbarBurgers = Array.prototype.slice.call(
    document.querySelectorAll(".navbar-burger"),
    0
  );
  // Check if there are any navbar burgers
  if ($navbarBurgers.length > 0) {
    // Add a click event on each of them
    $navbarBurgers.forEach((el) => {
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

  // blog page hero area slider start
  var sliderElement = document.getElementById("hero-area-blog");
  var interval = 0;

  function autoplay(run) {
    clearInterval(interval);
    interval = setInterval(() => {
      if (run && slider) {
        slider.next();
      }
    }, 4000);
  }

  var slider = new KeenSlider(sliderElement, {
    loop: true,
    duration: 2000,
    dragStart: () => {
      autoplay(false);
    },
    dragEnd: () => {
      autoplay(true);
    }
  });

  sliderElement?.addEventListener("mouseover", () => {
    autoplay(false);
  });
  sliderElement?.addEventListener("mouseout", () => {
    autoplay(true);
  });
  autoplay(true);
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
