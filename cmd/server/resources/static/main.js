let mainElement = document.getElementsByClassName("content")[0]
let dropArea = document.getElementById('uploader');


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

document.onpaste = function (event) {
    let items = (event.clipboardData || event.originalEvent.clipboardData).items;
    console.log(items)
    for (let index in items) {
        let item = items[index];
        if (item.kind === "file") {
            uploadFile(item.getAsFile())
        }
    }
};

function preventDefaults(e) {
    e.preventDefault()
    e.stopPropagation()
}

function highlight() {
    dropArea.classList.add('highlight')
}

function unhighlight() {
    dropArea.classList.remove('highlight')
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

function uploadFile(file) {
    let progressBar = document.createElement("progress")
    progressBar.setAttribute("max", "100")
    progressBar.setAttribute("value", "0")
    mainElement.appendChild(progressBar)
    let url = '/upload/file'
    let formData = new FormData()
    formData.append('file', file)

    function handleProgress(event, progressBar) {
        progressBar.value = event.loaded / event.total * 100
    }

    function progressDone(progressBar, response) {
        let leftText = document.createTextNode("Uploaded: ");
        let rightText = document.createTextNode(" (" + response.data.HumanSize + ")");
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
        .catch(() => { /* Error. Inform the user */})
}
