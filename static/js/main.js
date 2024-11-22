timezoneoffset = document.querySelectorAll("[name='timezoneoffset']")
currentTime = new Date()
for (let i = 0; i < timezoneoffset.length; i++) {
    timezoneoffset[i].value = currentTime.getTimezoneOffset()
}