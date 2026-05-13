import pytest
from django.conf import settings

@pytest.fixture(autouse=True)
def use_test_database():
    """اطمینان از استفاده از دیتابیس تست"""
    pass

@pytest.fixture(scope='session')
def django_db_setup(django_db_setup, django_db_blocker):
    """تنظیمات دیتابیس تست"""
    with django_db_blocker.unblock():
        pass