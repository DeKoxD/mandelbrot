// Source: https://medium.com/@jamischarles/what-is-debouncing-2505c0648ff1
function debounce(fn, time) {
	var timer;

	return function() {
		clearTimeout(timer);

		timer = setTimeout(() => {
			fn.apply(this, arguments);
		}, time);
	}
}

export default debounce;