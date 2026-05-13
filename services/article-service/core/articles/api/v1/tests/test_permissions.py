import pytest
from rest_framework import status, permissions
from django.urls import reverse
from ....models import Article
from .factories import ArticleFactory, UserFactory
from rest_framework.test import APIClient


pytestmark = pytest.mark.django_db

class TestIsAuthenticatedOrReadOnly:
    def test_read_operations_allow_unauthenticated(self, api_client):
        url = reverse('article:api-v1:article-list')
        response = api_client.get(url)
        assert response.status_code == status.HTTP_200_OK

    def test_write_operations_require_authentication(self, api_client, category, subcategory, tag):
        url = reverse('article:api-v1:article-list')
        data = {
            'title': 'Test',
            'read_time_min': 1,
            'short_description': 'desc',
            'content': 'content',
            'category': category.id,
            'subcategory': subcategory.id,
            'author': 'a',
            'identifier': 'i',
            'email': 'a@b.com',
            'tag': [tag.id],
            'article_image': 'http://example.com/img.jpg',
            'avatar': 'http://example.com/avatar.jpg',
        }
        response = api_client.post(url, data, format='json')
        assert response.status_code == status.HTTP_403_FORBIDDEN



