/* global $, WebSocket, location */

var musicplayerShowHide = function (self, playing) {
	if (playing) {
		$(self).find('.js-musicplayer-show-playing').show();
		$(self).find('.js-musicplyaer-show-stopped').hide();
	} else {
		$(self).find('.js-musicplayer-show-playing').hide();
		$(self).find('.js-musicplyaer-show-stopped').show();
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

var fireShowHide = function (self, on) {
	if (on) {
		$(self).find('.js-fire-show').show();
		$(self).find('.js-fire-on').hide();
	} else {
		$(self).find('.js-fire-show').hide();
		$(self).find('.js-fire-on').show();
	}
};

var relayShowHide = function (self, on) {
	if (on) {
		$(self).find('.js-relay-off').show();
		$(self).find('.js-relay-on').hide();
	} else {
		$(self).find('.js-relay-off').hide();
		$(self).find('.js-relay-on').show();
	}
};

var utrinekShowHide = function (self, on) {
	if (on) {
		$(self).find('.js-utrinek-show').show();
		$(self).find('.js-utrinek-on').hide();
	} else {
		$(self).find('.js-utrinek-show').hide();
		$(self).find('.js-utrinek-on').show();
	}
};

$(function () {
	$('.slider').slider();

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

		if ($(self).hasClass('js-musicplayer')) {
			musicplayerShowHide(self, $(self).find('input[name=playing]').val() === 'true');

			var socket = new WebSocket('ws://' + location.host + url + 'song');
			socket.onmessage = function (event) {
				$(self).find('.js-musicplayer-song').text(event.data);
			};

			$(self).find('select[name=playlist]').change(function () {
				var value = $(this).val();
				$.get(url + 'playlist/' + value);
			});

			$(self).find('.js-musicplayer-play').click(function () {
				$.get(url + 'play');
				musicplayerShowHide(self, true);
			});

			$(self).find('.js-musicplayer-stop').click(function () {
				$.get(url + 'stop');
				musicplayerShowHide(self, false);
			});

			$(self).find('.js-musicplayer-next').click(function () {
				$.get(url + 'next');
			});
		} else if ($(self).hasClass('js-fire')) {
			fireShowHide(self, $(self).find('input[name=on]').val() === 'true');

			$(self).find('.js-fire-on').click(function () {
				$.get(url + 'on');
				fireShowHide(self, true);
			});

			$(self).find('.js-fire-off').click(function () {
				$.get(url + 'off');
				fireShowHide(self, false);
			});

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
		} else if ($(self).hasClass('js-relay')) {
			relayShowHide(self, $(self).find('input[name=on]').val() === 'true');

			$(self).find('.js-relay-on').click(function () {
				$.get(url + 'on');
				relayShowHide(self, true);
			});

			$(self).find('.js-relay-off').click(function () {
				$.get(url + 'off');
				relayShowHide(self, false);
			});
		} else if ($(self).hasClass('js-utrinek')) {
			utrinekShowHide(self, $(self).find('input[name=on]').val() === 'true');

			$(self).find('.js-utrinek-on').click(function () {
				$.get(url + 'on');
				utrinekShowHide(self, true);
			});

			$(self).find('.js-utrinek-off').click(function () {
				$.get(url + 'off');
				utrinekShowHide(self, false);
			});

			$(self).find('.js-utrinek-go').click(function () {
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
