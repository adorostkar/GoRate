function validateForm()
{
    var x = document.settingsForm.extension.value
    try {
        new RegExp(x)
    } catch (e) {
        document.settingsForm.extension.focus()
        alert(x + " is not a valid regular expression")
        return false
    }
    x = document.settingsForm.separator.value
    try {
        new RegExp(x)
    } catch (e) {
        document.settingsForm.separator.focus()
        alert(x + " is not a valid regular expression")
        return false
    }
    x = document.settingsForm.nameparser.value.split("\n")
    var i = 0

    for (i = 0; i < x.length; i++) {
        try {
            console.log(x[i]);
            
            new RegExp(x[i])
        } catch (e) {
            document.settingsForm.nameparser.focus()
            alert(x[i] + " is not a valid regular expression")
            return false
        }
    }

    return ( true );
}