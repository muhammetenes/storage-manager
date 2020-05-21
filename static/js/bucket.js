$("#createBucketForm").submit(function () {
    var post_url = $(this).attr("action"); //get form action url
    var request_method = $(this).attr("method"); //get form GET/POST method
    var form_data = $(this).serialize(); //Encode form elements for submission
    var createBucketButton = $("#createBucketButton");
    $.ajax({
        url: post_url,
        type: request_method,
        data: form_data
    }).done(function (response) {
        $("#modalBody").text(response.message);
        if (response.error === false) {
            location.reload()
        }
        $("#createBucketModal").modal("hide");
        $("#infoModal").modal("show");
        createBucketButton.text("Create");
        $("#bucket_name").val("");
        createBucketButton.removeAttr("disabled")
    });
    return false
});
$("#createBucketButton").click(function () {
    var elem = $(this);
    elem.attr("disabled", "");
    elem.html("<span class=\"spinner-border spinner-border-sm\" role=\"status\" aria-hidden=\"true\"></span>\n" +
        "  Loading...");
    $("#createBucketForm").submit();
})

// Base delete bucket request
function deleteBucketsRequest(buckets) {
    return $.ajax({
        url: deleteBucketsUrl,
        method: "post",
        data: {
            buckets: buckets
        }
    })
}
// Single delete bucket
function deleteBucket(key, options) {
    var bucket_item = $(options["$trigger"][0]).parent();
    var bucketKey = options["$trigger"][0].dataset.caption;
    deleteBucketsRequest(bucketKey.split()).done(function (response) {
        $("#modalBody").text(response.message);
        $("#infoModal").modal("show");
        if (!response.error){
            bucket_item.remove();
        }
    });
}
// Multible delete buckets
function deleteBuckets(key, options) {
    var itemKeys = $(".bucket-item-checkbox:checked");
    var keys = [];
    var bucket_items = [];
    $.each(itemKeys, function (index, elem) {
        var bucket_item = $(elem).parent().parent();
        var key = bucket_item.children("a")[0].dataset.caption;
        keys.push(key);
        bucket_items.push(bucket_item)
    });
    deleteBucketsRequest(keys).done(function (response) {
        $("#modalBody").text(response.message);
        $("#infoModal").modal("show");
        if (!response.error){
            $.each(bucket_items, function (index, elem) {
                elem.remove();
            });
        }
    })
}

// CONTEXT MENU
$.contextMenu({
    selector: '.bucket-item',
    callback: function(key, options) {
        var m = "clicked: " + key;
        window.console && console.log(m) || alert(m);
    },
    items: {
        // "rename": {name: "Rename bucket", icon: "edit"},
        "delete": {name: "Delete bucket", icon: "delete", callback: deleteBucket},
        "sep1": "---------",
        // "duplicate": {name: "Duplicate bucket", icon: "copy"},
        "delete_selected": {name: "Delete selected buckets", icon:"delete", disabled: function () {
                return $(document).find(".bucket-item-checkbox:checked").length <= 1;
            }, callback: deleteBuckets},
    }
});