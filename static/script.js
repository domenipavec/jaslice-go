$(function() {
	$(".js-playlist-form").each(function() {
		var self = this;

		var url = $(this).find("input[name=url]").val();
		
		$(this).find("select[name=playlist]").change(function() {
			var value = $(this).val();
			$.get(url + "playlist/" + value);
		});
		
		$(this).find(".js-play").click(function() {
			$.get(url + "play");
			$(this).addClass("hidden");
			$(self).find(".js-stop").removeClass("hidden");
			$(self).find(".js-next").removeClass("hidden");
		});
		
		$(this).find(".js-stop").click(function() {
			$.get(url + "stop");
			$(this).addClass("hidden");
			$(self).find(".js-play").removeClass("hidden");
			$(self).find(".js-next").addClass("hidden");
		});	
	});
});
