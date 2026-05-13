import pytest
from rest_framework import status
from django.urls import reverse
from ....models import Comment

pytestmark = pytest.mark.django_db

class TestRemoteUserPayload:
    def test_comment_gets_username_from_mock_auth(self, authenticated_client, article):
        url = reverse('article:api-v1:comment-list')
        data = {'article': article.id, 'content': 'Hello'}
        response = authenticated_client.post(url, data, format='json')
        assert response.status_code == status.HTTP_201_CREATED
        comment = Comment.objects.first()
        assert comment.username == 'testuser'
        assert comment.citizenId == 'hm-2000003'
        assert comment.avatar == 'https://example.com/avatar.jpg'
        assert comment.slug_level == 5
        assert comment.image_level == 'https://example.com/level5.png'