var host = window.location.hostname + ":" + window.location.port

$(function(){
    $("tbody").find("tr").each(function () {
        var tr = $(this).children();
        var ws = new WebSocket("ws://"+host+"/progress");

        ws.onopen = function(evt) {
            ws.send(tr.eq(0).html());
        };

        ws.onmessage = function(evt) {

            var data = eval('(' + evt.data + ')');

            if (data.Status == 200) {
                ws.close()
                tr.eq(3).find(".progress-bar").width("100%")
                tr.eq(3).find(".progress-bar").html("100%")
                tr.eq(6).find("button").removeAttr("disabled")
                tr.eq(6).find("button").html("推送数据")
            } else {
                n = data.Progress/data.Total*100
                tr.eq(3).find(".progress-bar").width( n + "%")
                tr.eq(3).find(".progress-bar").html(n + "%")
                tr.eq(6).find("button").html("推送中")
            }

        };

        ws.onclose = function(evt) {
            console.log("Connection closed.");
        };


    })
});


function push(name, t) {
    $(t).attr('disabled', 'disabled')
    $(t).html("推送中")


    $.post('/push', { name: name }, function (text) {

        if (text.code != "200") {
            alert(text.message)
        }

    });


    setTimeout(function (name, t) {
        var ws = new WebSocket("ws://"+host+"/progress");
        ws.onopen = function(evt) {
            ws.send(name);
        };
        ws.onmessage = function(evt) {

            var data = eval('(' + evt.data + ')');

            if (data.Status == 200) {
                $(t).removeAttr("disabled")
                $(t).html("推送数据")
                ws.close()
                return
            } else {
                n = data.Progress/data.Total*100
                $(t).parent().parent().find(".progress-bar").width( n + "%")
                $(t).parent().parent().find(".progress-bar").html(n + "%")
            }


        };
        ws.onclose = function(evt) {
            console.log("Connection closed.");
        };
    },20000, name, t);

}