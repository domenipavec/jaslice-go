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

$(function () {
	modulesShowHide($('input[name=on]').val() === 'true');

	$('.js-on').click(function () {
		$.get('/api/on');
		modulesShowHide(true);
	});

	$('.js-off').click(function () {
		$.get('/api/off');
		modulesShowHide(false);
	});

	$('.js-module').each(function () {
		var url = $(this).find('input[name=url]').val();
		var self = this;

		if ($(this).hasClass('js-musicplayer')) {
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
		}
	});
});
