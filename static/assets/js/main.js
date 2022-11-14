document.addEventListener("DOMContentLoaded", () => {
  // AOS initialization
  AOS.init({
    once: true,
  });

  // Headroom js
  // var header = document.querySelector("header");
  // var headroom = new Headroom(header);
  // headroom.init();


  // navbar for mobile device
  let navbar = document.querySelector(".navbar-burger");
  navbar?.addEventListener("click", function () {
    const hasActiveClass = navbar.classList.contains("is-active");
    let dropdown = document.querySelector(".navbar-right");
    navbar.classList.toggle("is-active");
    dropdown.style.opacity = 1 - dropdown.style.opacity;
    dropdown.style.visibility = hasActiveClass ? "hidden" : "visible";
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
        window.scroll({
          top: 0,
          left: 0,
          behavior: "smooth",
        });
      }
    };
    // Listeners
    window.addEventListener("scroll", btnReveal);
    btnTop.addEventListener("click", TopscrollTo);
  };
  basicScrollTop();
  // TopscrollTo();
});


