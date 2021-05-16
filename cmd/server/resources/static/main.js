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
function preventDefaults (e) {
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
    let url = '/uploadfile'
    let formData = new FormData()
    formData.append('file', file)
    function handleProgress(event, progressBar) {
        progressBar.value = event.loaded / event.total * 100
    }
    function progressDone(progressBar, response) {
        console.log(response.data)
        let leftText = document.createTextNode("Uploaded: ");
        let rightText = document.createTextNode(" (" + response.data.HumanSize + ")");
        let link = document.createElement("a")
        link.href = response.data.URL
        link.innerHTML = response.data.FullName
        let para = document.createElement("p")
        console.log(leftText)
        console.log(link)
        console.log(rightText)
        para.appendChild(leftText)
        para.appendChild(link)
        para.appendChild(rightText)
        console.log(para)
        progressBar.replaceWith(para)
    }
    axios({
        method: 'post',
        url: url,
        data: formData,
        onUploadProgress: event => handleProgress(event, progressBar)
    })
        .then((response) => progressDone(progressBar, response))
        .catch(() => { /* Error. Inform the user */ })
}