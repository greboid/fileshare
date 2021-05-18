let dropArea = document.getElementById('uploader');
let deleteLinks = document.getElementsByClassName("deleteLink");

if (dropArea != null) {
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false)
    });
    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false)
    });
    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false)
    });
    dropArea.addEventListener('drop', handleDrop, false)
}

if (deleteLinks != null) {
    Array.from(deleteLinks).forEach(link => {
        link.addEventListener("click", () => handleDelete(link), false)
    })
}


function preventDefaults(e) {
    e.preventDefault()
    e.stopPropagation()
}

function highlight() {
    dropArea.firstElementChild.classList.add('highlight');
}

function unhighlight() {
    dropArea.firstElementChild.classList.remove('highlight');
}

function handleDrop(e) {
    let dt = e.dataTransfer
    let files = dt.files

    handleFiles(files)
}

function handleFiles(files) {
    ([...files]).forEach(file => {
        uploadFile(file)
    })
}

function handleDelete(link) {
    axios({
        withCredentials: true,
        method: 'delete',
        url: '/admin/delete/' + link.dataset.file,
    })
        .then(() => link.parentNode.parentNode.removeChild(link.parentNode))
        .catch(() => {
            /* Error. Inform the user */
        })
}

function uploadFile(file) {
    let mainElement = document.getElementsByClassName("uploaded")[0]
    let expiryElement = document.getElementById('expiry');
    let randomiseElement = document.getElementById('randomise');
    let expiry = expiryElement.selectedOptions[0].value
    let progressBar = document.createElement("progress")
    progressBar.setAttribute("max", "100")
    progressBar.setAttribute("value", "0")
    mainElement.appendChild(progressBar)
    let url = '/upload/file'
    let formData = new FormData()
    formData.append('file', file)
    formData.append('expiry', expiry)
    formData.append('randomise', randomiseElement.checked)

    function handleProgress(event, progressBar) {
        progressBar.value = event.loaded / event.total * 100
    }

    function progressDone(progressBar, response) {
        let leftText = document.createTextNode("Uploaded: ");
        let rightText = document.createTextNode(" (" + response.data.HumanSize + ") - Expires: " + response.data.HumanExpiry);
        let link = document.createElement("a")
        link.href = response.data.URL
        link.innerHTML = response.data.FullName
        let para = document.createElement("p")
        para.appendChild(leftText)
        para.appendChild(link)
        para.appendChild(rightText)
        progressBar.replaceWith(para)
    }

    axios({
        withCredentials: true,
        method: 'post',
        url: url,
        data: formData,
        onUploadProgress: event => handleProgress(event, progressBar)
    })
        .then((response) => progressDone(progressBar, response))
        .catch(() => {
            /* Error. Inform the user */
        })
}
