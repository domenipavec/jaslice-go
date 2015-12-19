/* global $, WebSocket, location */

$(function () {
	$('.slider').slider();

	var generalShowHide = function (self, on) {
		if (on) {
			$(self).find('.js-module-show-on').show();
			$(self).find('.js-module-off').show();
			$(self).find('.js-module-show-off').hide();
			$(self).find('.js-module-on').hide();
		} else {
			$(self).find('.js-module-show-on').hide();
			$(self).find('.js-module-off').hide();
			$(self).find('.js-module-show-off').show();
			$(self).find('.js-module-on').show();
		}
	};

	var modulesShowHide = function (on) {
		if (on) {
			$('.js-module').show();
			$('.js-on').hide();
			$('.js-off').show();
		} else {
			$('.js-module').hide();
			$('.js-musicplayer').show();
			$('.js-on').show();
			$('.js-off').hide();
		}
	};

	modulesShowHide($('input[name=on]').val() === 'true');

	$('.js-on').click(function () {
		$.get('/api/on');

		// everything gets reset on turn on, so this is the easiest solution
		setTimeout(function () {
			location.reload();
		}, 500);
	});

	$('.js-off').click(function () {
		$.get('/api/off');
		modulesShowHide(false);
	});

	$('.js-module').each(function () {
		var url = $(this).find('input[name=url]').val();
		var self = this;

		if ($(self).find('input[name=on]').length) {
			generalShowHide(self, $(self).find('input[name=on]').val() === 'true');

			$(self).find('.js-module-on').click(function () {
				$.get(url + 'on');
				generalShowHide(self, true);
			});

			$(self).find('.js-module-off').click(function () {
				$.get(url + 'off');
				generalShowHide(self, false);
			});
		}

		if ($(self).hasClass('js-musicplayer')) {
			var socket = new WebSocket('ws://' + location.host + url + 'song');
			socket.onmessage = function (event) {
				$(self).find('.js-musicplayer-song').text(event.data);
			};

			$(self).find('select[name=playlist]').change(function () {
				var value = $(this).val();
				$.get(url + 'playlist/' + value);
			});

			$(self).find('.js-musicplayer-next').click(function () {
				$.get(url + 'next');
			});
		} else if ($(self).hasClass('js-fire')) {
			$(self).find('input[name=color]').on('slideStop', function () {
				var value = $(this).val();
				$.get(url + 'color/' + value);
			});

			$(self).find('input[name=light]').on('slideStop', function () {
				var value = $(this).val();
				$.get(url + 'light/' + value);
			});

			$(self).find('input[name=speed]').on('slideStop', function () {
				var value = $(this).val();
				$.get(url + 'speed/' + value);
			});
		} else if ($(self).hasClass('js-nebo')) {
			$(self).find('select[name=mode]').change(function () {
				var value = $(this).val();
				$.get(url + 'mode/' + value);
			});

			$(self).find('input[name=speed]').on('slideStop', function () {
				var value = $(this).val();
				$.get(url + 'speed/' + value);
			});
		} else if ($(self).hasClass('js-pwm')) {
			$(self).find('input[name=value]').on('slideStop', function () {
				var value = $(this).val();
				$.get(url + '/' + value);
			});
		} else if ($(self).hasClass('js-utrinek')) {
			$(self).find('.js-utrinek-show').click(function () {
				$.get(url + 'show');
			});

			var oldVal = $(self).find('input[name=interval]').val().split(',');
			$(self).find('input[name=interval]').on('slideStop', function () {
				var newVal = $(this).val().split(',');
				if (newVal[0] !== oldVal[0]) {
					$.get(url + 'min/' + newVal[0]);
				}
				if (newVal[1] !== oldVal[1]) {
					$.get(url + 'max/' + newVal[1]);
				}
				oldVal = newVal;
			});
		}
	});
});
