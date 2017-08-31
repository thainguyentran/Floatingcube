openerp.epos_foundation_display_customer_name_on_receipt = function (instance) {
	var module = instance.epos_foundation;
	var QWeb = instance.web.qweb;
	var _t = instance.web._t;

	module.receiptFormats.include({
		init: function (posmodel, options) {
			this.esc_com = new module.EscCommand();
			this.pos = posmodel;
		},

		_print_receipt_header: function (json) {
			var line = [];
			line = line.concat(this.esc_com.draw_at_left("Receipt No: " + json.name));
			line = line.concat(this.esc_com.draw_at_left("Cashier: " + json.cashier));
			if (json.client)
			{
				line = line.concat(this.esc_com.draw_at_left("Client Name: " + json.client.name));
			}

			line = line.concat(this.esc_com.draw_at_left(json.date.localestring));

			if (json.floor && json.table)
			{
				line = line.concat(this.esc_com.draw_at_left("At floor/table: " + json.floor + "/" + json.table));
			}

			if (json.buzzer)
			{
				line = line.concat(this.esc_com.draw_at_left("buzzer : " + json.buzzer));
			}
			
			if (json.customer_count)
			{
				line = line.concat(this.esc_com.draw_at_left("Pax Number : " + json.customer_count));
			}

			line = line.concat(this.esc_com.draw_at_left(this.esc_com.get_dash_line()));

			return line;
		},
	});
};

