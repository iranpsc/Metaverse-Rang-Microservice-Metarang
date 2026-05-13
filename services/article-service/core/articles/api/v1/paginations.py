from rest_framework.pagination import PageNumberPagination    
from rest_framework.response import Response


class ArticlePagination(PageNumberPagination):
      page_size = 10
      def get_paginated_response(self, data):
        return Response({
            'links': {
                'next': self.get_next_link(),
                'previous': self.get_previous_link()
            },
            'total_object': self.page.paginator.count,
            'total_page': self.page.paginator.num_pages,
            'results': data
        })