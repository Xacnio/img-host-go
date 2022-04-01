var UPLOAD = false;

function UploadStatus() {
    if(UPLOAD === true) {
        if(!$(".upload-select").hasClass("ready")) $(".upload-select").addClass("ready")
        if($(".upload-button").is(":hidden")) $(".upload-button").fadeIn()
    } else {
        if($(".upload-select").hasClass("ready")) $(".upload-select").removeClass("ready")
        if($(".upload-button").is(":visible")) $(".upload-button").hide()
    }
}

window.onload = function () {
    $("#images").change(function() {
        document.getElementById("files").innerHTML = "";
        if(this.files.length > 8) {
            alert("You can upload max 8 imagess!")
            document.getElementById("images").value = "";
            return
        }
        if(this.files.length === 0) {
            if(UPLOAD === true) {
                UPLOAD = false;
                UploadStatus()
            }
        }
        [].forEach.call(this.files, function(file, index){
            if(file.size > 5 * 1024 * 1024) {
                if(UPLOAD === true) {
                    UPLOAD = false;
                    UploadStatus()
                }
                document.getElementById("files").innerHTML = "";
                document.getElementById("images").value = "";
                alert("Max file size is 5 MB!")
                return
            }
            var reader = new FileReader();
            reader.onload = (function(theFile) {
                return function(e) {
                    let fileName = file.name;
                    if(fileName.length > 20) {
                        const ext = fileName.split(/[\s.]+/)
                        const extension = ext[ext.length - 1]
                        const fileNameS = fileName.split('.').slice(0, -1).join('.');
                        fileName = fileNameS.substring(0, 16) + "." + extension
                    }
                    var div = document.createElement('div');
                    div.className = 'col-sm-12';
                    div.innerHTML = '<img src="'+e.target.result+'" /> ' + fileName + " (" + niceBytes(file.size) + ")";
                    document.getElementById("files").append(div);
                };
            })(file);
            reader.readAsDataURL(file);
            if(UPLOAD === false) {
                UPLOAD = true;
                UploadStatus()
            }
        });
    });
}

function copyText2(eleme)
{
    var textArea = document.getElementById(eleme);
    textArea.select();

    try
    {
        var successful = document.execCommand( 'copy' );
        var msg = successful ? 'successful' : 'unsuccessful';
        //console.log('Copying text command was ' + msg);
    }
    catch (err) {
        console.log('Oops, unable to copy',err);
    }
}

function niceBytes(x){
    const units = ['bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    let l = 0, n = parseInt(x, 10) || 0;
    while(n >= 1024 && ++l){
        n = n/1024;
    }
    return(n.toFixed(n < 10 && l > 0 ? 1 : 0) + ' ' + units[l]);
}

$(function() {
    $('img.lazy').Lazy({
        effect: 'fadeIn',
        effectTime: 250,
    });
});