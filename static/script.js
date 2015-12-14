$(function() {
	$(".js-playlist-form").each(function() {
		var self = this;

		var url = $(this).find("input[name=url]").val();

		var socket = new WebSocket("ws://" + location.host + url + "song");
		socket.onmessage = function (event) {
			$(self).find(".js-current-song").text(event.data);
		}

		$(this).find("select[name=playlist]").change(function() {
			var value = $(this).val();
			$.get(url + "playlist/" + value);
		});

		$(this).find(".js-play").click(function() {
			$.get(url + "play");
			$(this).addClass("hidden");
			$(self).find(".js-stop").removeClass("hidden");
			$(self).find(".js-next").removeClass("hidden");
			$(self).find(".js-song").removeClass("hidden");
		});

		$(this).find(".js-stop").click(function() {
			$.get(url + "stop");
			$(this).addClass("hidden");
			$(self).find(".js-play").removeClass("hidden");
			$(self).find(".js-next").addClass("hidden");
			$(self).find(".js-song").addClass("hidden");
		});

		$(this).find(".js-next").click(function() {
			$.get(url + "next");
		});
	});
});
