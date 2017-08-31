

_columns = {
'email_opt' : fields.related('config_id', 'email_session_report', type='boolean', relation='pos.config', string="Email Session Report Option"),
}

_defaults = {
        'name' : '/',
        'user_id' : lambda obj, cr, uid, context: uid,
        'state' : 'opening_control',
        'sequence_number': 1,
        'login_number': 0,
        'email_opt': False,
    }
    
def send_report_email(self, cr, uid, ids, context=None):
        email_template_obj = self.pool.get('mail.template')
        pos_session_obj = self.pool.get('pos.print.session.summary')
        pos_order_obj = self.pool.get('pos.order')
        session = self.browse(cr, uid, ids)
        report_data = pos_session_obj.get_print_session_summary_report_json(cr, uid, session.id, context=None)
        order_data = pos_order_obj.get_session_info(cr, uid, session.id, None, context=None)
        report_header = report_data['header']
        email_body = ""
        lines = []
        line_size_limit = 50
        template_ids = email_template_obj.search(cr, uid, [('name', '=', 'End Of Session Report')], context=context)
        if template_ids:
            values = email_template_obj.generate_email(cr, uid, template_ids[0], ids[0], context=context)
            #Generate the Content of the email
            #POS Session Summary Report
            #header
            email_body += "POS Session Summary Report<br>"
            email_body += "( {} )<br>".format(report_data['session_name'])
            email_body += "Generated on {} <br>".format(report_data['now'])
            email_body += "{}<br>".format(report_header['company_name'])
            email_body += "POS: {}<br>".format(report_header['config_name'])
            email_body += "--------------------------------------<br>"
            #body
            email_body += "Sale Statistics<br>"
            email_body += "-------------------------------------------<br>"

            for payment in report_data['payments']:
                method_name = payment['name']
                total = payment['sum']
                count = payment['count']
                part1 = "{0} {1} Sales".format(count, method_name)
                part2 = '{:.>50}<br>'.format(total)
                lines.append([part1, part2])
            lines.append(["{0} Voided Sales".format(report_data['voided_count']), '{:.>50}<br>'.format(report_data['voided_total'])])
            lines.append(['-' * line_size_limit + '<br>'])

            lines.append(["TOTAL DISCOUNT", '{:.>50}<br>'.format(report_data['total_discount'])])
            lines.append(["TOTAL NET SALES", '{:.>50}<br>'.format(report_data['net_sale'])])
            lines.append(["TOTAL TAXES", '{:.>50}<br>'.format(report_data['total_tax'])])
            lines.append(["TOTAL SALES", '{:.>50}<br>'.format(report_data['payment_total'])])
            lines.append(["<br>Other Statistics<br>"])
            lines.append(['-' * line_size_limit + '<br>'])
            #calucation
            invoice_average = report_data['payment_total'] / report_data['payment_count']
            average_item_per_invoice = report_data['orders_item_count'] / report_data['payment_count']
            average_item_price = report_data['payment_total'] / report_data['orders_item_count']
            #report strings
            lines.append(["{0} Invoice Average".format(report_data['payment_count']), '{:.>50}<br>'.format(invoice_average)])
            lines.append(["Average items per invoice", '{:.>50}<br>'.format(average_item_per_invoice)])
            lines.append(["{0} items average price".format(report_data['orders_item_count']), '{:.>50}<br>'.format(average_item_price)])
            lines.append(["Number of customers", '{:.>50}<br>'.format(report_data['customer_count'])])
            lines.append(["Average spend per customers", '{:.>50}<br>'.format(report_data['average_spend'])])
            lines.append(["<br>Sales by Category<br>"])
            lines.append(["--------<br>"])
            category_sales = report_data['categ_sales']
            for cat_name, cat_total in category_sales.items():
                lines.append(["{0}".format(cat_name), '{:.>50}<br>'.format(cat_total)])

            for line_parts in lines:
                if len(line_parts) > 1:
                    to_remove = (len(line_parts[0]) + len(line_parts[1])) - line_size_limit
                    if to_remove > 0:
                        new_part2 = line_parts[1][to_remove:]
                        email_body += line_parts[0] + new_part2
                    else:
                        email_body += line_parts[0] + line_parts[1]
                else: email_body += line_parts[0]

            email_body += "<br>=============================================================<br>"
            #Sale by product report
            email_body += "<br>SALE BY PRODUCT REPORT<br><br>"
            orders = order_data['orders']
            all_orders = []
            add_new = True
            for order in orders:
                order_detail = order['order_detail']
                for order_line in order_detail:
                    for single_order in all_orders:
                        if single_order['name'] == order_line['name']:
                            single_order['qty'] += order_line['qty']
                            single_order['total'] += order_line['total']
                            add_new = False
                            break
                        else: add_new = True
                    if add_new:
                        each_order = {}
                        each_order['name'] = order_line['name']
                        each_order['qty'] = order_line['qty']
                        each_order['price_unit'] = order_line['price_unit']
                        each_order['total'] = order_line['total']
                        all_orders.append(each_order)
            lines = []
            collum1 = len("Product" + '{:.>40}'.format("Quantity"))
            collum2 = len("Quantity" + '{:.>20}'.format("Unit Price"))
            collum3 = len("Unit Price" + '{:.>20}<br>'.format("Subtotal"))
            header = "Product" + '{:.>40}'.format("Quantity") + '{:.>20}'.format("Unit Price") + '{:.>20}<br>'.format("Subtotal")
            line_size_limit = len(header)
            email_body += header
            for order_total in all_orders:
                lines.append([order_total['name'], '{:.>50}'.format(order_total['qty']), '{:.>20}'.format(order_total['price_unit']), '{:.>20}<br>'.format(order_total['total'])])
            for line_parts in lines:
                to_remove1 = len(line_parts[0] + line_parts[1]) - collum1
                if to_remove1 > 0:
                    new_part1 = line_parts[1][to_remove1:]
                else: new_part1 = line_parts[1]
                to_remove2 = len(line_parts[1].replace('.', '') + line_parts[2]) - collum2
                if to_remove2 > 0:
                    new_part2 = line_parts[2][to_remove2:]
                else: new_part2 = line_parts[2]
                to_remove3 = len(line_parts[2].replace('.', '') + line_parts[3]) - collum3
                if to_remove3 > 0:
                    new_part3 = line_parts[3][to_remove3:]
                else: new_part3 = line_parts[3]
                email_body += line_parts[0] + new_part1 + new_part2 + new_part3
            email_body += "======================================<br>"
            #Contents of Email
            values['email_from'] = "EPOSAdmin@floatingcube.com"
            values['body_html'] = '<font face="Courier New">' + email_body +'</font>'
            values['res_id'] = False
            mail_mail_obj = self.pool.get('mail.mail')
            msg_id = mail_mail_obj.create(cr, uid, values, context=context)
            if msg_id:
                mail_mail_obj.send(cr, uid, [msg_id], context=context)
