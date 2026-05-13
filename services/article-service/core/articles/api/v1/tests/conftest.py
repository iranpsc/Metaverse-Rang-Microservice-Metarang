import pytest
from django.core.files.uploadedfile import SimpleUploadedFile
from django.contrib.auth import get_user_model
from rest_framework.test import APIClient
from PIL import Image
import io
from unittest.mock import patch
from django.test import override_settings
from ....models import Category, SubCategory, Tag, Article, Comment, Reply
from .factories import (
    UserFactory, CategoryFactory, SubCategoryFactory, TagFactory,
    ArticleFactory, CommentFactory, ReplyFactory
)
from rest_framework.authentication import BaseAuthentication
from django.contrib.auth.models import User
from django.conf import settings



User = get_user_model()

class MockJWTAuthentication(BaseAuthentication):
    def authenticate(self, request):
        auth_header = request.headers.get('Authorization', '')
        if not auth_header.startswith('Bearer '):
            return None
        
        user, _ = User.objects.get_or_create(username='testuser')
        request.remote_user_payload = {
            'username': 'testuser',
            'citizenId': 'hm-2000003',   # ادمین واقعی
            'avatar': 'https://example.com/avatar.jpg',
            'slug_level': 5,
            'image_level': 'https://example.com/level5.png',
        }
        return (user, None)

@pytest.fixture(autouse=True, scope='session')
def mock_auth_settings():
    with override_settings(REST_FRAMEWORK={
        'DEFAULT_AUTHENTICATION_CLASSES': [
            'articles.api.v1.tests.conftest.MockJWTAuthentication',
        ]
    }):
        yield

@pytest.fixture
def api_client():
    return APIClient()

@pytest.fixture
def authenticated_client(api_client):
    api_client.credentials(HTTP_AUTHORIZATION='Bearer any-token')
    return api_client

@pytest.fixture
def admin_client(api_client):
    api_client.credentials(HTTP_AUTHORIZATION='Bearer admin-token')
    return api_client

@pytest.fixture
def another_user(db):
    return UserFactory(username='another', email='another@test.com')

@pytest.fixture
def test_user(db):
    return UserFactory()


@pytest.fixture
def test_admin_user(db):
    return UserFactory(is_staff=True, is_superuser=True)


@pytest.fixture
def sample_image():
    image = Image.new('RGB', (100, 100), color='red')
    image_io = io.BytesIO()
    image.save(image_io, format='JPEG')
    image_io.seek(0)
    return SimpleUploadedFile('test_image.jpg', image_io.read(), content_type='image/jpeg')

@pytest.fixture
def invalid_image():
    return SimpleUploadedFile('test.txt', b'not an image', content_type='text/plain')

@pytest.fixture
def category(db):
    return CategoryFactory()

@pytest.fixture
def subcategory(db, category):
    return SubCategoryFactory(category=category)

@pytest.fixture
def tag(db):
    return TagFactory()

@pytest.fixture
def article(db, category, subcategory, tag):
    article = ArticleFactory(category=category, subcategory=subcategory)
    article.tag.add(tag)
    return article

@pytest.fixture
def comment(db, article):
    return CommentFactory(article=article)

@pytest.fixture
def reply(db, comment):
    return ReplyFactory(comment=comment)

@pytest.fixture
def mock_remote_auth_success():
    with patch('requests.get') as mock_get:
        mock_response = mock_get.return_value
        mock_response.status_code = 200
        mock_response.json.return_value = {
            'data': {
                'name': 'remoteuser',
                'code': '1234567890',
                'image': 'http://example.com/avatar.jpg',
                'level': 2,
            }
        }
        yield mock_get

@pytest.fixture
def mock_remote_auth_failure():
    with patch('requests.get') as mock_get:
        from requests.exceptions import HTTPError
        mock_response = mock_get.return_value
        mock_response.status_code = 401
        mock_response.raise_for_status.side_effect = HTTPError("401 Client Error", response=mock_response)
        mock_response.json.return_value = {'error': 'Invalid token'}
        yield mock_get