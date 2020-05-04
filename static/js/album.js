// File upload
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


// Create Folder
$("#createFolderButton").click(function () {
    var elem = $(this);
    elem.attr("disabled", "");
    elem.html("<span class=\"spinner-border spinner-border-sm\" role=\"status\" aria-hidden=\"true\"></span>\n" +
        "  Loading...");
    $("#createFolderForm").submit();
})
$("#createFolderForm").submit(function () {
    var post_url = $(this).attr("action"); //get form action url
    var request_method = $(this).attr("method"); //get form GET/POST method
    var form_data = new FormData(this);
    var createFolderButton = $("#createFolderButton");
    $.ajax({
        url: post_url,
        type: request_method,
        data: form_data,
        processData: false,  // Important!
        contentType: false,
        cache: false,
    }).done(function (response) {
        $("#modalBody").text(response.message);
        if (response.error === false) {
            location.reload()
        }
        $("#createFolderModal").modal("hide");
        $("#exampleModalCenter").modal("show");
        createFolderButton.text("Upload");
        $("#bucket_name").val("");
        createFolderButton.removeAttr("disabled")
    });
    return false
});



// ------ PAGINATION ------
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
// ------ PAGINATION-END ------


// ------ PHOTO ------
function deleteItemsRequest(keys) {
    return $.ajax({
        url: deleteItemUrl,
        method: "post",
        data: {
            keys: keys
        }
    })
}

function deleteItem(key, options) {
    var photo_item = $(options["$trigger"][0]).parent();
    var itemKey = options["$trigger"][0].dataset.caption;
    deleteItemsRequest(itemKey.split()).done(function (response) {
        $("#modalBody").text(response.message);
        $("#exampleModalCenter").modal("show");
        if (!response.error){
            photo_item.remove();
            item_count--
        }
        $(".object-count").text(item_count)
    });
}

function deleteItems(key, options) {
    var itemKeys = $(".photo-item-checkbox:checked");
    var keys = [];
    var photo_items = [];
    $.each(itemKeys, function (index, elem) {
        var photo_item = $(elem).parent().parent();
        var key = photo_item.children("a")[0].dataset.caption;
        keys.push(key);
        photo_items.push(photo_item)
    });
    deleteItemsRequest(keys).done(function (response) {
        $("#modalBody").text(response.message);
        $("#exampleModalCenter").modal("show");
        if (!response.error){
            $.each(photo_items, function (index, elem) {
                elem.remove();
            });
            AOS.init({
                duration: 800,
                easing: 'slide',
                once: false
            });
            item_count -= photo_items.length
        }
        $(".object-count").text(item_count)
    })
}

// Photo context menu
var download_item = $("#download-item")
$.contextMenu({
    selector: '.album-item',
    callback: function(key, options) {
        var m = "clicked: " + key;
        window.console && console.log(m) || alert(m);
    },
    items: {
        // "rename": {name: "Rename", icon: "edit"},
        // "move": {name: "Move", icon: "paste"},
        // "copy": {name: "Copy", icon: "copy"},
        "delete": {name: "Delete", icon: "delete", callback: deleteItem},
        "download": {name: "Download", icon: "download", callback: function (key, options) {
                download_item.attr("href", options.$trigger[0].dataset.url);
                download_item[0].click()
            }},
        "sep1": "---------",
        // "move_selected": {name: "Move selected items", icon: "paste", disabled: function () {
        // 		return $(document).find(".photo-item-checkbox:checked").length <= 1;
        // 	}},
        // "copy_selected": {name: "Copy selected items", icon: "copy", disabled: function () {
        // 		return $(document).find(".photo-item-checkbox:checked").length <= 1;
        // 	}},
        "delete_selected": {name: "Delete selected items", icon:"delete", disabled: function () {
                return $(document).find(".photo-item-checkbox:checked").length <= 1;
            }, callback: deleteItems},
        "download_selected": {name: "Download selected items", icon: "download", disabled: function () {
                return $(document).find(".photo-item-checkbox:checked").length <= 1;
            }, callback:function (key, options) {
                $.each($(document).find(".photo-item-checkbox:checked"), function (i, elem) {
                    var data_url = $(elem).parent().parent().children("a")[0].dataset.url
                    download_item.attr("href", data_url);
                    download_item[0].click()
                })
            }}
    }
});

// ------ PHOTO-END ------

// ------ FOLDER ------
function deleteFoldersRequest(keys) {
    return $.ajax({
        url: deleteFolderUrl,
        method: "post",
        data: {
            keys: keys
        }
    })
}

function deleteFolder(key, options) {
    var photo_item = $(options["$trigger"][0]).parent();
    var itemKey = options["$trigger"][0].dataset.caption;
    deleteFoldersRequest(itemKey.split()).done(function (response) {
        $("#modalBody").text(response.message);
        $("#exampleModalCenter").modal("show");
        if (!response.error){
            photo_item.remove();
            item_count--
        }
        $(".object-count").text(item_count)
    });
}


// Folder context menu
$.contextMenu({
    selector: '.folder-item',
    callback: function(key, options) {
        var m = "clicked: " + key;
        window.console && console.log(m) || alert(m);
    },
    items: {
        "delete": {name: "Delete", icon: "delete", callback: deleteFolder}
    }
});

// ------ FOLDER-END ------