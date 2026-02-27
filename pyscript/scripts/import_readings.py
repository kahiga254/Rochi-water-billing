import pandas as pd
import numpy as np
from datetime import datetime
import re
from pymongo import MongoClient

# MongoDB connection
client = MongoClient('mongodb+srv://adamskahiga:36596768Bantu.@cluster0.9wu5s9u.mongodb.net/')
db = client['water_billing']

def clean_phone(phone):
    """Clean and format phone numbers to +254 format"""
    if pd.isna(phone) or str(phone).strip() in ['', ' ', '0.0']:
        return ''
    
    phone = str(phone).strip()
    # Remove non-digits
    phone = re.sub(r'\D', '', phone)
    
    if not phone or phone == '':
        return ''
    
    # Format to +254...
    if phone.startswith('0'):
        phone = '254' + phone[1:]
    elif phone.startswith('7'):
        phone = '254' + phone
    elif phone.startswith('254') and len(phone) >= 9:
        pass
    else:
        return ''
    
    return '+' + phone

def import_kagerema_january_readings():
    """Import Kagerema January readings as last readings"""
    
    print("\n📊 Processing Kagerema January readings...")
    
    # Kagerema January data
    kagerema_readings = [
        {"name": "Martin Gichere", "meter": "20210304039", "current": 1591, "previous": 1487},
        {"name": "Josphat Kariuki", "meter": "22050367", "current": 451, "previous": 426},
        {"name": "Jane Wairegi", "meter": "22051234", "current": 38, "previous": 38},
        {"name": "Phillip Msee", "meter": "22050315", "current": 527, "previous": 504},
        {"name": "Andrew Gatonye", "meter": "22051446", "current": 847, "previous": 743},
        {"name": "Eshter (Miriam tenant)", "meter": "15081262", "current": 229, "previous": 229},
        {"name": "Salome Ngugi(flats)", "meter": "22051207", "current": 1072, "previous": 1008},
        {"name": "Lucy Muraguri", "meter": "20230601541", "current": 290, "previous": 278},
        {"name": "Benard Omwando", "meter": "20210300841", "current": 150, "previous": 139},
        {"name": "Eunice Wahome", "meter": "22050592", "current": 134, "previous": 128},
        {"name": "Patrick Kamami", "meter": "15080477", "current": 168, "previous": 158},
        {"name": "Paul Wahinya", "meter": "15081171", "current": 40, "previous": 0},
        {"name": "Michael Ndung'u", "meter": "23030546", "current": 31, "previous": 29},
        {"name": "Teresia gitau", "meter": "23041306", "current": 116, "previous": 81},
        {"name": "Simon wakaba", "meter": "20210302885", "current": 113, "previous": 104},
        {"name": "Irene Kinyanjui", "meter": "20230601631", "current": 85, "previous": 76},
        {"name": "Salome Wanjiru", "meter": "20230603677", "current": 432, "previous": 394},
        {"name": "Janefaith", "meter": "20230604373", "current": 0, "previous": 0},
        {"name": "Mt Zion", "meter": "20230603714", "current": 256, "previous": 240},
        {"name": "Migwi", "meter": "23050676", "current": 59, "previous": 58},
        {"name": "Mary Igweta", "meter": "240805871", "current": 126, "previous": 121},
        {"name": "Kestine Karimi", "meter": "20240804834", "current": 5, "previous": 5},
        {"name": "Francisca Kimani", "meter": "20240805865", "current": 151, "previous": 136},
        {"name": "Lucy macharia", "meter": "20240603871", "current": 92, "previous": 73},
        {"name": "Paul Karanja", "meter": "20241005643", "current": 113, "previous": 77},
        {"name": "Mary Ndekei", "meter": "20241008129", "current": 62, "previous": 46},
        {"name": "William Kimanthi", "meter": "20240630275", "current": 197, "previous": 180},
        {"name": "Esther Muthoni", "meter": "20240120403", "current": 99, "previous": 87},
        {"name": "Steve H", "meter": "20241008175", "current": 6, "previous": 0},
        {"name": "Edward Wandaka", "meter": "20420810905", "current": 32, "previous": 29},
        {"name": "Festus Maina", "meter": "20240811700", "current": 51, "previous": 43},
        {"name": "John Kathuri Njeru", "meter": "20240813157", "current": 26, "previous": 19},
        {"name": "Euticus Wangai", "meter": "20240810902", "current": 24, "previous": 23},
        {"name": "Edna Ochako", "meter": "20240701058", "current": 54, "previous": 44},
        {"name": "Komu M mboche", "meter": "20240703294", "current": 103, "previous": 58},
        {"name": "John Njeru", "meter": "20241201169", "current": 49, "previous": 40},
        {"name": "Titus Muriithi", "meter": "20240812351", "current": 81, "previous": 50},
        {"name": "Bernad M Mungai", "meter": "20241200668", "current": 137, "previous": 112},
        {"name": "Obadia G. Ethaiba", "meter": "20241200795", "current": 216, "previous": 187},
        {"name": "Simon M Ngugi", "meter": "20240707524", "current": 59, "previous": 43},
        {"name": "Thomas Ndug'u", "meter": "20240703980", "current": 30, "previous": 23},
        {"name": "Jane Gathuru", "meter": "20240701387", "current": 10, "previous": 9},
        {"name": "Nicholas Muita", "meter": "20241234629", "current": 29, "previous": 17},
        {"name": "Nancy Matianyi", "meter": "20241130494", "current": 11, "previous": 3},
        {"name": "Joseph muriuki", "meter": "20241101571", "current": 37, "previous": 12},
        {"name": "Isaac Nthiga", "meter": "20241235078", "current": 19, "previous": 10},
        {"name": "Wajogoo", "meter": "", "current": 10, "previous": 0},
        {"name": "Salome Ng'ang'a", "meter": "20250912210", "current": 2, "previous": 0},
    ]
    
    return process_readings(kagerema_readings, "Kagerema", "January 2026")

def import_mwana_january_readings():
    """Import Mwana Wikio January readings as last readings"""
    
    print("\n📊 Processing Mwana Wikio January readings...")
    
    # Mwana Wikio January data
    mwana_readings = [
        {"name": "Harriet Wanjiku Methu", "meter": "20240631593", "current": 253, "previous": 211},
        {"name": "Martha Ngugi", "meter": "20240809692", "current": 33, "previous": 29},
        {"name": "Omwango Evans", "meter": "20240809605", "current": 50, "previous": 43},
        {"name": "Peter G.Gichure", "meter": "20240809754", "current": 145, "previous": 121},
        {"name": "Ben Mutua", "meter": "20240809628", "current": 107, "previous": 94},
        {"name": "Miriam Mwangi", "meter": "20240811176", "current": 21, "previous": 20},
        {"name": "Samuel W.Njoroge", "meter": "20240813476", "current": 69, "previous": 56},
        {"name": "Muriithi .B", "meter": "20240813429", "current": 21, "previous": 21},
        {"name": "Priscah wangui", "meter": "20240812341", "current": 97, "previous": 74},
        {"name": "Edwin Kahiga", "meter": "20240803294", "current": 26, "previous": 21},
        {"name": "Christine Ngugi", "meter": "20241104679", "current": 36, "previous": 30},
        {"name": "Benson Kariuki", "meter": "20240704822", "current": 30, "previous": 20},
        {"name": "Rechael Njuguna", "meter": "20241101436", "current": 7, "previous": 7},
        {"name": "Edward Ndegwa", "meter": "20241234030", "current": 14, "previous": 6},
        {"name": "Jane Kimbe", "meter": "20241200871", "current": 6, "previous": 0},
        {"name": "Philip Musyoka", "meter": "20241105204", "current": 0, "previous": 0},
    ]
    
    return process_readings(mwana_readings, "Mwana Wikio", "January 2026")

def process_readings(reading_list, zone, period):
    """Process readings and update customer last_reading"""
    
    customers_collection = db['customers']
    readings_collection = db['meter_readings']
    
    updated = 0
    not_found = 0
    skipped = 0
    readings_created = 0
    
    for reading in reading_list:
        meter = reading['meter']
        
        # Skip if no meter number
        if not meter or meter == '':
            print(f"  ⚠️ Skipping {reading['name']} - no meter number")
            skipped += 1
            continue
        
        # Find the customer
        customer = customers_collection.find_one({'meter_number': meter})
        
        if not customer:
            print(f"  ❌ Customer not found for meter: {meter} ({reading['name']})")
            not_found += 1
            continue
        
        # Create a reading date for January 2026
        reading_date = datetime(2026, 1, 1)
        
        # Calculate consumption
        consumption = reading['current'] - reading['previous']
        if consumption < 0:
            consumption = 0
        
        # Create a meter reading record (optional - for history)
        reading_doc = {
            'meter_number': meter,
            'customer_id': customer['_id'],
            'customer_name': reading['name'],
            'reading_date': reading_date,
            'previous_reading': reading['previous'],
            'current_reading': reading['current'],
            'consumption': consumption,
            'rate_per_unit': 100.0,
            'water_charge': consumption * 100,
            'fixed_charge': 0.0,
            'reading_type': 'manual',
            'reading_method': 'import',
            'month': 'JANUARY',
            'year': 2026,
            'billing_period': period,
            'status': 'recorded',
            'created_at': datetime.now()
        }
        
        readings_collection.insert_one(reading_doc)
        readings_created += 1
        
        # Update customer with last reading
        result = customers_collection.update_one(
            {'_id': customer['_id']},
            {
                '$set': {
                    'last_reading': reading['current'],
                    'last_reading_date': reading_date,
                    'updated_at': datetime.now()
                }
            }
        )
        
        if result.modified_count > 0:
            updated += 1
            print(f"  ✅ Updated {reading['name']}: last_reading = {reading['current']} (prev: {reading['previous']})")
    
    return {
        'updated': updated,
        'not_found': not_found,
        'skipped': skipped,
        'readings_created': readings_created
    }

def main():
    print("=" * 60)
    print("🚀 IMPORT JANUARY READINGS AS LAST READINGS")
    print("=" * 60)
    print("This script will:")
    print("  1. Set January readings as customer last_reading")
    print("  2. Store readings in meter_readings collection")
    print("  3. NOT create any bills")
    print("=" * 60)
    
    # Process Kagerema
    kag_results = import_kagerema_january_readings()
    
    # Process Mwana Wikio
    mwana_results = import_mwana_january_readings()
    
    # Combine results
    total_updated = kag_results['updated'] + mwana_results['updated']
    total_not_found = kag_results['not_found'] + mwana_results['not_found']
    total_skipped = kag_results['skipped'] + mwana_results['skipped']
    total_readings = kag_results['readings_created'] + mwana_results['readings_created']
    
    print("\n" + "=" * 60)
    print("📊 IMPORT SUMMARY")
    print("=" * 60)
    print(f"✅ Customers updated with last_reading: {total_updated}")
    print(f"📊 Meter readings created: {total_readings}")
    print(f"⚠️ Customers not found: {total_not_found}")
    print(f"⏩ Skipped (no meter): {total_skipped}")
    print("=" * 60)
    
    # Show sample of updated customers
    print("\n📋 Sample customers with new last readings:")
    for customer in db.customers.find({'last_reading': {'$gt': 0}}).limit(5):
        print(f"  • {customer['first_name']} {customer['last_name']}: {customer['last_reading']} m³")

if __name__ == "__main__":
    main()