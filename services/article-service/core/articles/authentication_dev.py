
# dev
# from rest_framework.authentication import BaseAuthentication
# from rest_framework.exceptions import AuthenticationFailed
# from .mock_auth import MockJWTAuthenticationBackend

# class MockJWTAuthentication(BaseAuthentication):
#     def authenticate(self, request):
#         auth_header = request.headers.get('Authorization', '')
#         if not auth_header.startswith('Bearer '):
#             return None
        
#         backend = MockJWTAuthenticationBackend()
#         user = backend.authenticate(request, token=None)  # token不重要
#         if user is None:
#             raise AuthenticationFailed('Invalid token.')
        
#         return (user, None)



from rest_framework.authentication import BaseAuthentication
from rest_framework.exceptions import AuthenticationFailed
from .mock_auth import MockJWTAuthenticationBackend

class MockJWTAuthentication(BaseAuthentication):
    def authenticate(self, request):
        auth_header = request.headers.get('Authorization', '')
        if not auth_header.startswith('Bearer '):
            return None

        backend = MockJWTAuthenticationBackend()
        user = backend.authenticate(request, token=auth_header.split(' ')[1])
        if user is None:
            raise AuthenticationFailed('Invalid token.')
        return (user, None)