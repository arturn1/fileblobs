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
  
  // Verificar se estamos na raiz do projeto padrão
  const urlParams = new URLSearchParams(window.location.search);
  const prefix = urlParams.get("prefix") || "";
  
  // Checar se temos folderActionButtons (pode não existir na raiz da conta padrão)
  const folderActionButtons = document.getElementById("folderActionButtons");
});

function showDownloadConfirm() {
  isDownloadMode = true;

  // Verificar se os elementos existem antes de manipulá-los
  const folderActionButtons = document.getElementById("folderActionButtons");
  const confirmButtons = document.getElementById("confirmButtons");
  
  if (folderActionButtons) folderActionButtons.style.display = "none";
  if (confirmButtons) confirmButtons.style.display = "flex";

  document.querySelectorAll(".file").forEach(el => {
    el.classList.add("show-checkboxes");
  });

  document.querySelectorAll(".folder").forEach(folder => {
    folder.classList.add("download-folder");
  });
}

function cancelDownload() {
  isDownloadMode = false;

  // Verificar se os elementos existem antes de manipulá-los
  const folderActionButtons = document.getElementById("folderActionButtons");
  const confirmButtons = document.getElementById("confirmButtons");
  
  if (confirmButtons) confirmButtons.style.display = "none";
  if (folderActionButtons) folderActionButtons.style.display = "flex";

  document.querySelectorAll(".file").forEach(el => {
    el.classList.remove("show-checkboxes", "selected");
    el.querySelector(".file-checkbox").checked = false;
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

function toggleFileSelection(el, event) {
  const checkbox = el.querySelector(".file-checkbox");

  if (isDownloadMode) {
    event.preventDefault();
    checkbox.checked = !checkbox.checked;
    el.classList.toggle("selected", checkbox.checked);
  } else {
    const path = el.getAttribute("data-path");
    window.location.href = "/download?path=" + encodeURIComponent(path);
  }
}


function selectAll() {
  document.querySelectorAll(".file-checkbox").forEach(cb => {
    cb.checked = true;
    cb.closest(".file").classList.add("selected");
  });
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

  const urlParams = new URLSearchParams(window.location.search);
  const prefix = urlParams.get("prefix") || "";
  const prefixInput = document.createElement("input");
  prefixInput.type = "hidden";
  prefixInput.name = "prefix";
  prefixInput.value = prefix;
  form.appendChild(prefixInput);

  document.body.appendChild(form);
  form.submit();
}
