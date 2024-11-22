function notify(title, title2, message) {
    document.querySelector("#notify-title").innerText = title
    document.querySelector("#notify-title2").innerText = title2
    document.querySelector("#notify-message").innerText = message
    const toastLiveExample = document.getElementById('liveToast')
    const toastBootstrap = bootstrap.Toast.getOrCreateInstance(toastLiveExample)
    toastBootstrap.show()
}


function formatDateTime(date, format="Y-m-d H:M:s") {
    var d = new Date(date)

    result = ""
    for (var i = 0; i < format.length; i++) {
        switch (format[i]) {
            case "Y":
                result = result + d.getFullYear()
                break;
            case "m":
                result += preZero(d.getMonth(), 2)
                break;
            case "d":
                result += preZero(d.getDate(), 2)
                break;
            case "H":
                result += preZero(d.getHours(), 2)
                break;
            case "M":
                result += preZero(d.getMinutes(), 2)
                break;
            case "s":
                result += preZero(d.getSeconds(), 2)
                break;

            default:
                result += format[i]
        }
    }

    return result
}

function preZero(v, l) {
    v = v.toString()
    while (v.length < l) {
        v = "0"+v
    }
    return v
}