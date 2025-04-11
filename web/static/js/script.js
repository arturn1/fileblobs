let isDownloadMode = false;

document.addEventListener("DOMContentLoaded", () => {
  const searchInput = document.querySelector("#searchInput");

  if (searchInput) {
    searchInput.addEventListener("input", () => {
      const query = searchInput.value.toLowerCase();
      document.querySelectorAll(".folder, .file").forEach(card => {
        const text = card.textContent.toLowerCase();
        card.style.display = text.includes(query) ? "block" : "none";
      });
    });
  }

  document.addEventListener("click", function (e) {
    if (!isDownloadMode) return;
    const fileCard = e.target.closest(".file");
    const isLink = e.target.closest("a");
    if (fileCard && !isLink) {
      e.preventDefault();
      const checkbox = fileCard.querySelector(".file-checkbox");
      if (checkbox) {
        checkbox.checked = !checkbox.checked;
        fileCard.classList.toggle("selected", checkbox.checked);
      }
    }
  });
});

function showDownloadConfirm() {
  isDownloadMode = true;

  document.getElementById("actionButtons").style.display = "none";
  document.getElementById("confirmButtons").style.display = "flex";

  document.querySelectorAll(".file").forEach(el => {
    el.classList.add("show-checkboxes");
    el.querySelector(".file-link")?.removeAttribute("href");
  });

  document.querySelectorAll(".folder").forEach(folder => {
    folder.classList.add("download-folder");
  });
}

function cancelDownload() {
  isDownloadMode = false;

  document.getElementById("confirmButtons").style.display = "none";
  document.getElementById("actionButtons").style.display = "flex";

  document.querySelectorAll(".file").forEach(el => {
    el.classList.remove("show-checkboxes", "selected");
    el.querySelector(".file-checkbox").checked = false;
    const path = el.getAttribute("data-path");
    el.querySelector(".file-link")?.setAttribute("href", "/download?path=" + path);
  });

  document.querySelectorAll(".folder").forEach(folder => {
    folder.classList.remove("download-folder");
  });
}

function handleFolderClick(el) {
  const path = el.getAttribute("data-path");
  if (!isDownloadMode) {
    window.location.href = "/?prefix=" + path + "/";
  } else {
    window.location.href = "/download-folder?path=" + path;
  }
}

function downloadSelected() {
  const selected = Array.from(document.querySelectorAll(".file-checkbox:checked"))
    .map(cb => cb.value);

  if (selected.length === 0) {
    alert("Selecione ao menos 1 arquivo.");
    return;
  }

  const form = document.createElement("form");
  form.method = "POST";
  form.action = "/download-multiple";

  selected.forEach(path => {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = "files";
    input.value = path;
    form.appendChild(input);
  });

  document.body.appendChild(form);
  form.submit();
}
