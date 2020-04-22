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
        $("#exampleModalCenter").modal("show");
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