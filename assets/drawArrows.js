var cvs = document.getElementsByClassName("arrows")

for (cts of cvs) {
    cts.width = 10;
    cts.height = 22;

    var ctx = cts.getContext('2d');
    ctx.fillStyle = '#ccc'

    ctx.beginPath();
    ctx.moveTo(0, 10);
    ctx.lineTo(5, 0);
    ctx.lineTo(10, 10);
    ctx.fill();

    ctx.beginPath();
    ctx.moveTo(0, 12);
    ctx.lineTo(10, 12);
    ctx.lineTo(5, 22);
    ctx.fill();
}
