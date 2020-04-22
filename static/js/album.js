$("#uploadFileForm").submit(function () {
    var post_url = $(this).attr("action"); //get form action url
    var request_method = $(this).attr("method"); //get form GET/POST method
    var form_data = new FormData(this);
    var uploadFileButton = $("#uploadFileButton");
    $.ajax({
        url: post_url,
        type: request_method,
        data: form_data,
        enctype: 'multipart/form-data',
        processData: false,  // Important!
        contentType: false,
        cache: false,
    }).done(function (response) {
        $("#modalBody").text(response.message);
        if (response.error === false) {
            location.reload()
        }
        $("#UploadFileModal").modal("hide");
        $("#exampleModalCenter").modal("show");
        uploadFileButton.text("Upload");
        $("#bucket_name").val("");
        uploadFileButton.removeAttr("disabled")
    });
    return false
});
$("#uploadFileButton").click(function () {
    var elem = $(this);
    elem.attr("disabled", "");
    elem.html("<span class=\"spinner-border spinner-border-sm\" role=\"status\" aria-hidden=\"true\"></span>\n" +
        "  Loading...");
    $("#uploadFileForm").submit();
})

function getObjectsBeforeKey(key) {
    url = nextObjectsUrl + "?last_key=" + key + "&folder_key=" + folder_key
    is_request_run = true
    return $.ajax({
        url: url,
        method: "get",
    })
}

is_request_run = false
$(window).scroll(function () {
    if ((($(window).scrollTop() + $(window).height()) >= ($(document).height() - 500)) && !is_request_run) {
        var last_key = $(".album-item").last()[0].dataset.caption
        getObjectsBeforeKey(last_key).done(function (response) {
            if (response.objects.length > 0) {
                $.each(response.objects, function (i, obj) {
                    if (obj.Type) {
                        var photo_item = '<div class="col-6 col-md-6 col-lg-3" data-aos="fade-up">' +
                            '<div class="checkbox-btn"><input type="checkbox" class="photo-item-checkbox"/></div>' +
                            '<a href="' + obj.Url + '" data-url="' + obj.Url + '" class="d-block photo-item album-item" data-fancybox="gallery" data-caption="' + obj.Name + '">' +
                            '<img src="' + obj.Url + '" alt="Image" class="img-fluid">' +
                            '<div class="photo-text-more">' +
                            '<div class="photo-text-more">' +
                            '<h3 class="heading">' + obj.Name + '</h3>' +
                            '<br><span class="icon icon-search"></span>' +
                            '</div>' +
                            '</div>' +
                            '</a>' +
                            '</div>';
                    } else {
                        var photo_item = '<div class="col-6 col-md-6 col-lg-3" data-aos="fade-up">' +
                            '<div class="checkbox-btn"><input type="checkbox" class="photo-item-checkbox"/></div>' +
                            '<a href="/static/file_icon.png" data-url="' + obj.Url + '" class="d-block photo-item album-item" data-fancybox="gallery" data-caption="' + obj.Name + '">' +
                            '<img src="/static/file_icon.png" alt="Image" class="img-fluid">' +
                            '<div class="photo-text-more">' +
                            '<div class="photo-text-more">' +
                            '<h3 class="heading">' + obj.Name + '</h3>' +
                            '<br><span class="icon icon-search"></span>' +
                            '</div>' +
                            '</div>' +
                            '</a>' +
                            '</div>';
                    }
                    $(".align-items-stretch").append(photo_item)
                })
                AOS.init({
                    duration: 800,
                    easing: 'slide',
                    once: false
                });
                is_request_run = false
            }
        })
    }
})