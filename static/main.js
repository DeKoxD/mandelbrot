import debounce from './debounce.js';

document.addEventListener('DOMContentLoaded', event => {
  var canvas = document.querySelector("#fractal");
  var ctx = canvas.getContext('2d');
  Object.assign(canvas.dataset, {
    centerx: -0.5,
    centery: 0,
    zoom: 200,
    resx: canvas.offsetWidth,
    resy: canvas.offsetHeight,
    offsetx: 0,
    offsety: 0,
    mousedown: false,
  });

  var draw = loadCanvasFractal(canvas, 1.5, 300);

  canvas.addEventListener('onresize', () => {
    Object.assign(canvas.dataset, {
      resx: canvas.offsetWidth,
      resy: canvas.offsetHeight
    });
    draw();
  });
  var zoom = (amount) => {
    return (amt) => {
      canvas.dataset.zoom = parseFloat(canvas.dataset.zoom, 10) * (amount ? amount : amt);
      draw(true);
    };
  };
  document.querySelector("#zoom-in").onclick = zoom(1.25);
  document.querySelector("#zoom-out").onclick = zoom(1/1.25);
  canvas.onmousewheel = (() => {
    var zoomfunc = zoom();
    var inc = 1.25;
    return ((e) => {
      zoomfunc(e.deltaY > 0 ? 1/inc : inc);
    });
  })();
  canvas.onmousedown = () => {
    if(canvas.dataset.mousedown === "false") {
      canvas.dataset.mousedown = true;
    }
  };
  canvas.onmouseup = () => {
    if(canvas.dataset.mousedown === "true") {
      canvas.dataset.mousedown = false;
      canvas.dataset.centerx = parseFloat(canvas.dataset.centerx, 10) -
        (canvas.dataset.offsetx/canvas.dataset.zoom + 1/(2*canvas.dataset.zoom));
      canvas.dataset.centery = parseFloat(canvas.dataset.centery, 10) +
        (canvas.dataset.offsety/canvas.dataset.zoom + 1/(2*canvas.dataset.zoom));
      canvas.dataset.offsetx = 0;
      canvas.dataset.offsety = 0;
      draw(true);
    }
  };
  canvas.onmousemove = (e) => {
    if(canvas.dataset.mousedown === "true") {
      canvas.dataset.offsetx = parseInt(canvas.dataset.offsetx, 10) + e.movementX;
      canvas.dataset.offsety = parseInt(canvas.dataset.offsety, 10) + e.movementY;
      draw(false);
    }
  };
  draw(true);
});

function loadCanvasFractal(canvas, extra, time) {
  var img;
  const fetchFunc = debounce(async (ctx, centerx, centery, zoom, resxextra, resyextra, offsetx, offsety, resx, resy) => {
    const raw = await fetchFractal(centerx, centery, zoom, resxextra, resyextra, true, false);
    img = new ImageData(resxextra, resyextra);
    for(let i = 0, j = 0; i < img.data.length; i += 4, j++) {
      img.data[i + 0] = raw[j] ? 0 : 255;
      img.data[i + 1] = raw[j] ? 0 : 255;
      img.data[i + 2] = raw[j] ? 0 : 255;
      img.data[i + 3] = 255;
    }
    ctx.putImageData(img, offsetx-Math.ceil((resxextra-resx)/2), offsety-Math.ceil((resyextra-resy)/2));
  }, time);
  return async function(refetch) {
    const { offsetx, offsety } = canvas.dataset;
    const { centerx, centery, zoom, resx, resy } = canvas.dataset;
    const [resxextra, resyextra] = [Math.ceil(resx*extra), Math.ceil(resy*extra)];
    const ctx = canvas.getContext('2d');
    if (refetch || !img) {
      fetchFunc(ctx, centerx, centery, zoom, resxextra, resyextra, offsetx, offsety, resx, resy);
    } else {
      ctx.putImageData(img, offsetx-Math.ceil((resxextra-resx)/2), offsety-Math.ceil((resyextra-resy)/2));
    }
  }
};

async function fetchFractal(centerx, centery, zoom, resx, resy, trueArg, falseArg) {
  var url = new URL(window.location.href + 'fractal');
  var params = {
    centerx,
    centery,
    zoom,
    resx,
    resy,
  };
  Object.keys(params).forEach(key => url.searchParams.append(key, params[key]));

  var result = await fetch(url, { method:'GET', });
  result = await result.json();
  return base64bin2bool(result.Image, trueArg, falseArg).slice(0, result.ResX*result.ResY);
}

function bin2bool(input) {
  let arr = [];
  for(let i = 0; i < 8; i++) {
    arr.push(!!(input&(1 << i)));
  }
  return arr;
};

function base64bin2bool(input, trueArg, falseArg) {
  var trueVal = trueArg === null ? true : trueArg;
  var falseVal = falseArg === null ? false : falseArg;
  var img = [];
  let encImg = atob(input);
  for (let i = 0; i < encImg.length;i++) {
    img.push(
      ...(bin2bool( encImg.charCodeAt(i)).reduce(
        (acc, item) => acc.concat(item ? trueVal : falseVal), []
      ))
    );
  }
  return img;
};