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

document.addEventListener("DOMContentLoaded", () => {
  const allHeaders = document.querySelectorAll(
    ".blog-content > h2,.blog-content > h3,.blog-content > h4"
  );
  const blogThumb = document.querySelector(".blog-content .fig-style");

  // docs page header link create
  Array.from(allHeaders).forEach((el) => {
    const id = el.id;
    const anchorTag = document.createElement("a");
    anchorTag.setAttribute("href", "#" + id);
    anchorTag.innerHTML = `<svg width="25" height="13" viewBox="0 0 52.965 52.973">
    <g id="broken-link" transform="translate(-0.004)">
      <path id="Path_16124" data-name="Path 16124" d="M49.467,3.51a12.011,12.011,0,0,0-16.97,0L23.305,12.7a1,1,0,0,0,1.414,1.414l9.192-9.192A10,10,0,0,1,48.052,19.066L36.033,31.088a10.014,10.014,0,0,1-14.143,0A1,1,0,0,0,20.476,32.5a12.013,12.013,0,0,0,16.97,0L49.467,20.48a12.03,12.03,0,0,0,0-16.97Z" fill="#4a4a4a"/>
      <path id="Path_16125" data-name="Path 16125" d="M26.84,40.279l-7.778,7.778A10,10,0,1,1,4.92,33.915L16.234,22.6a10.015,10.015,0,0,1,14.142,0,1,1,0,0,0,1.414-1.414,12.011,12.011,0,0,0-16.97,0L3.505,32.5A11.987,11.987,0,0,0,11.99,52.973a11.911,11.911,0,0,0,8.485-3.5l7.778-7.778a1,1,0,1,0-1.413-1.414Z" fill="#4a4a4a"/>
      <path id="Path_16126" data-name="Path 16126" d="M33.969,44.009a1,1,0,0,0-1,1v6a1,1,0,0,0,2,0v-6A1,1,0,0,0,33.969,44.009Z" fill="#4a4a4a"/>
      <path id="Path_16127" data-name="Path 16127" d="M38.433,42.3a1,1,0,0,0-1.414,1.414l4.243,4.242a1,1,0,0,0,1.414-1.414Z" fill="#4a4a4a"/>
      <path id="Path_16128" data-name="Path 16128" d="M44.969,38.009h-6a1,1,0,0,0,0,2h6a1,1,0,0,0,0-2Z" fill="#4a4a4a"/>
      <path id="Path_16129" data-name="Path 16129" d="M15.969,11.009a1,1,0,0,0,1-1v-6a1,1,0,1,0-2,0v6A1,1,0,0,0,15.969,11.009Z" fill="#4a4a4a"/>
      <path id="Path_16130" data-name="Path 16130" d="M11.5,12.716A1,1,0,0,0,12.918,11.3L8.676,7.06A1,1,0,0,0,7.262,8.474Z" fill="#4a4a4a"/>
      <path id="Path_16131" data-name="Path 16131" d="M4.969,17.009h6a1,1,0,0,0,0-2h-6a1,1,0,0,0,0,2Z" fill="#4a4a4a"/>
    </g>
   </svg>`;
    el.appendChild(anchorTag);

    //insert hash tag when click anchorTag
    anchorTag.addEventListener("click", (e) => {
      e.preventDefault();
      const targetEl = document.querySelector(e.currentTarget.hash);
      window.history.pushState(id, "title", "#" + id);
      const pos1 = targetEl.offsetTop + (blogThumb.clientHeight - 40);

      window.scrollTo({
        top: pos1,
        behavior: "smooth",
      });
    });
  });

  //docs page heading content on reload
  setTimeout(function () {
    let getHash = location.hash;
    if (getHash) {
      const targetE2 = document.querySelector(getHash);
      const pos2 = targetE2.offsetTop + (blogThumb.clientHeight - 40);
      scrollTo({
        top: pos2,
        behavior: "smooth",
      });
    }
  }, 0);
});
