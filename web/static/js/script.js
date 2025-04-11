console.log("Hello, World!");

document.addEventListener("DOMContentLoaded", function () {
    const searchInput = document.querySelector("#searchInput");
    const folderCards = document.querySelectorAll(".folder");
    const fileCards = document.querySelectorAll(".file");
  
    if (!searchInput) return;
  
    const updateVisibility = (cards, query) => {
      cards.forEach(card => {
        const text = card.textContent.toLowerCase();
        card.style.display = text.includes(query) ? "block" : "none";
      });
    };
  
    searchInput.addEventListener("input", () => {
      const query = searchInput.value.toLowerCase();
      updateVisibility(folderCards, query);
      updateVisibility(fileCards, query);
    });
  });
  