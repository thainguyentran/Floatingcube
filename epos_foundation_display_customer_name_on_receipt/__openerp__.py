# -*- coding: utf-8 -*-
{
    'name': "Display Customer Name On Receipt",

    'summary': """
        Display the customer's name on receipt""",

    'description': """
        Need to install epos_foundation first
    """,

    'author': "Floating Cube Studios",
    'website': "http://www.epos.com.sg/",

    # Categories can be used to filter modules in modules listing
    # Check https://github.com/odoo/odoo/blob/master/openerp/addons/base/module/module_data.xml
    # for the full list
    'category': 'Point Of Sale',
    'version': '0.1',

    # any module necessary for this one to work correctly
    'depends': ['base', 'epos_foundation'],

    # always loaded
    'data': [
        'templates.xml',
    ],
}