from pymongo import MongoClient
from datetime import datetime

# MongoDB connection
client = MongoClient('mongodb+srv://adamskahiga:36596768Bantu.@cluster0.9wu5s9u.mongodb.net/')
db = client['water_billing']

def reset_database():
    """Delete ALL data from the database - customers, readings, bills"""
    
    print("⚠️  ⚠️  ⚠️  WARNING ⚠️  ⚠️  ⚠️ ")
    print("This will DELETE ALL data from the database:")
    print("  • All customers")
    print("  • All meter readings")
    print("  • All bills")
    print("  • All payments")
    print("\nThe system will show ZERO customers, ZERO bills, ZERO money collected.")
    print("\nPress Ctrl+C to cancel")
    print("Or type 'DELETE EVERYTHING' to proceed:")
    
    confirmation = input()
    
    if confirmation != "DELETE EVERYTHING":
        print("\n❌ Operation cancelled.")
        return
    
    print("\n🗑️  Resetting database...")
    print("=" * 40)
    
    # Delete all documents from each collection
    collections = ['customers', 'meter_readings', 'bills', 'payments']
    total_deleted = 0
    
    for collection_name in collections:
        collection = db[collection_name]
        
        # Count before deletion
        count_before = collection.count_documents({})
        
        # Delete all
        result = collection.delete_many({})
        
        print(f"✅ Cleared {collection_name}: {result.deleted_count} documents deleted (was {count_before})")
        total_deleted += result.deleted_count
    
    print("=" * 40)
    print(f"\n✨ Database reset complete!")
    print(f"📊 Total documents deleted: {total_deleted}")
    print("\n🏁 The system now has:")
    print("   • 0 customers")
    print("   • 0 bills")
    print("   • 0 readings")
    print("   • 0 payments")
    print("   • 0 money collected")
    
    # Final verification
    print("\n🔍 Verifying...")
    for collection_name in collections:
        count = db[collection_name].count_documents({})
        print(f"  {collection_name}: {count} documents")

if __name__ == "__main__":
    reset_database()