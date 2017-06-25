var arrowEls = document.getElementsByClassName("arrows");
var sts = sortTimestamp(true)
var sid = sortID(false)
var sun = sortUsername(false)

// sort by timstamp initially
v.rows = sts(v.rows)
// add click handlers
for (arrowEl of arrowEls) {
    arrowEl.addEventListener("click", sort)
}
// handle all the sort conditions
function sort(evt) {
    var el = evt.target
    if (el.id === "canvas-ts") {
        v.rows = sts(v.rows)
    } else if (el.id === "canvas-id") {
        v.rows = sid(v.rows)
    } else if (el.id === "canvas-un") {
        v.rows = sun(v.rows)
    }
}

// init_order is a boolean
// false -> lowest to highest
// true -> highest to lowest
function sortTimestamp(init_order){
    var sort_order = init_order
    return function(arr) {
        var sorted = arr.sort(function(a,b) {
            if (sort_order) {
                return new Date(b.Timestamp) - new Date(a.Timestamp)
            } else {
                return new Date(a.Timestamp) - new Date(b.Timestamp)
            }
        })
        sort_order = !sort_order
        return sorted
    }
}

function sortID(init_order){
    var sort_order = init_order
    return function(arr) {
        var sorted = arr.sort(function(a,b) {
            if (sort_order) {
                return a.ID > b.ID
            } else {
                return a.ID < b.ID
            }
        })
        sort_order = !sort_order
        return sorted
    }
}

function sortUsername(init_order){
    var sort_order = init_order
    return function(arr) {
        var sorted =  arr.sort(function(a,b) {
            if (sort_order) {
                return a.UserName.localeCompare(b.UserName)
            } else {
                return b.UserName.localeCompare(a.UserName)
            }
        })
        sort_order = !sort_order
        return sorted
    }
}
