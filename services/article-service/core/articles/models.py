from django.urls import reverse
from django.core.validators import RegexValidator
from django.core.exceptions import ValidationError
from django.db.models import Q
from django_jalali.db import models as jmodels
from django.db import transaction, IntegrityError, models
import re



class Article(models.Model):
    title = models.CharField(max_length=255) 
    slug = models.CharField(max_length=255, validators=[RegexValidator(r'^[a-zA-Z0-9-]*$', 'Slug only allows English characters, numbers, and hyphens.')], unique=True)
    read_time_min = models.PositiveIntegerField()
    article_image = models.ImageField(upload_to='article-images/')
    short_description = models.CharField(max_length=255)
    content = models.TextField()

    category = models.ForeignKey('Category', on_delete=models.PROTECT, null=False, blank=False)
    subcategory = models.ForeignKey('SubCategory', on_delete=models.PROTECT, null=False, blank=False) 

    author = models.CharField(max_length=255)
    identifier = models.CharField(max_length=255)
    avatar = models.ImageField(default='avatar-images/24eaa929-c37b-329a-80f2-5a72bd552305_rSRZN8L.jpg', upload_to='avatar-images/')
    field_of_activity = models.CharField(max_length=255, null=True, blank=True)
    biography = models.CharField(max_length=255, null=True, blank=True)
    telegram_id = models.CharField(max_length=255, null=True, blank=True)
    whatsapp_number = models.CharField(max_length=11, null=True, blank=True)  
    email = models.EmailField()

    tag = models.ManyToManyField('Tag', related_name='articles', blank=False)



    created_date = jmodels.jDateTimeField(auto_now_add=True)
    updated_date = jmodels.jDateTimeField(auto_now=True)


    class Meta:
        ordering = ['-created_date']




    def save(self, *args, **kwargs):
        if not self.pk:
            max_retries = 10
            
            pattern = re.compile(r'^bl-(\d+)$')
            max_num = 1000
            for article in Article.objects.only('slug'):
                match = pattern.match(article.slug)
                if match:
                    num = int(match.group(1))
                    if num > max_num:
                        max_num = num
            next_num = max_num + 1

            for attempt in range(max_retries):
                self.slug = f"bl-{next_num}"
                try:
                    with transaction.atomic():
                        super().save(*args, **kwargs)
                        return
                except IntegrityError:
                    next_num += 1
                    continue
            raise IntegrityError("Unable to generate unique slug after multiple attempts")
        else:
            super().save(*args, **kwargs)


    def get_absolute_api_url(self):
        return reverse("article:api-v1:article-detail", kwargs={"pk": self.pk})



    
    def get_snippet(self):
        if len(self.content) > 150:
            return self.content[:150] + " ..."
        return self.content



    def __str__(self):
        return self.title






class Category(models.Model):
    name = models.CharField(max_length=255, unique=True)
    category_slug = models.CharField(max_length=255, validators=[RegexValidator(r'^[a-zA-Z0-9-]*$', 'Slug only allows English characters, numbers, and hyphens.')])
    category_image = models.ImageField(upload_to='category-image/')
    category_description = models.CharField(max_length=255)
 

    def __str__(self):
        return self.name
    


class SubCategory(models.Model):
    category = models.ForeignKey('Category', on_delete=models.CASCADE, null=False, blank=False, related_name="subcategories")
    name = models.CharField(max_length=255, unique=True)

    def __str__(self):
        return self.name
    




class Tag(models.Model):
    label = models.CharField(max_length=250)
    slug = models.CharField(max_length=250, unique=True)


    def __str__(self):
            return self.label
    




class Comment(models.Model):
    article = models.ForeignKey("Article", on_delete=models.CASCADE, related_name='comments', null=False, blank=False)
    username = models.CharField(max_length=255)
    citizenId = models.CharField(max_length=255)
    avatar = models.URLField(max_length=500) 
    slug_level = models.PositiveIntegerField()
    image_level = models.URLField(max_length=500)
    content = models.TextField()
    approved = models.BooleanField(default=True) 
    created_date = jmodels.jDateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-created_date']



    def __str__(self):
        return f'Comment by {self.username} on {self.article.title}'
    




class Reply(models.Model):
    comment = models.ForeignKey('Comment', on_delete=models.CASCADE, related_name='replies')
    username = models.CharField(max_length=255)
    citizenId = models.CharField(max_length=255)
    avatar = models.URLField(max_length=500) 
    slug_level = models.PositiveIntegerField()
    image_level = models.URLField(max_length=500)
    content = models.TextField()
    approved = models.BooleanField(default=True) 
    created_date = jmodels.jDateTimeField(auto_now_add=True)

    class Meta:
        ordering = ['-created_date']

    def __str__(self):
        return f'Reply by {self.username} on {self.comment}'





class LikeDislike(models.Model):

    LIKE = 1
    DISLIKE = -1
    
    VOTES = (
        (LIKE, 'Like'),
        (DISLIKE, 'Dislike'),
    )

    vote = models.SmallIntegerField(choices=VOTES)
    

    username = models.CharField(max_length=255)
    citizenId = models.CharField(max_length=255)

    article = models.ForeignKey('Article', on_delete=models.CASCADE, related_name='likes', null=True, blank=True)
    comment = models.ForeignKey('Comment', on_delete=models.CASCADE, related_name='likes', null=True, blank=True)
    reply = models.ForeignKey('Reply', on_delete=models.CASCADE, related_name='likes', null=True, blank=True)
    
    created_date = jmodels.jDateTimeField(auto_now_add=True)

    class Meta:
        constraints = [

            models.UniqueConstraint(
                fields=['citizenId', 'article'],
                name='unique_user_article_vote'
            ),

            models.UniqueConstraint(
                fields=['citizenId', 'comment'],
                name='unique_user_comment_vote'
            ),

            models.UniqueConstraint(
                fields=['citizenId', 'reply'],
                name='unique_user_reply_vote'
            ),

            models.CheckConstraint(
                condition=(
                    (
                        Q(article__isnull=False) &
                        Q(comment__isnull=True) &
                        Q(reply__isnull=True)
                    ) |
                    (
                        Q(article__isnull=True) &
                        Q(comment__isnull=False) &
                        Q(reply__isnull=True)
                    ) |
                    (
                        Q(article__isnull=True) &
                        Q(comment__isnull=True) &
                        Q(reply__isnull=False)
                    )
                ),
                name='exactly_one_target'
            )
        ]


    def clean(self):
        selected = sum([bool(self.article), bool(self.comment), bool(self.reply)])

        if selected != 1:
            raise ValidationError("Exactly one target required.")


    def save(self, *args, **kwargs):
        self.clean()
        super().save(*args, **kwargs)



    def __str__(self):
        item_str = ""
        if self.article:
            item_str = f"Article '{self.article.title}'"
        elif self.comment:
            item_str = f"Comment on '{self.comment.article.title}'"
        elif self.reply:
            item_str = f"Reply on Comment ID {self.reply.comment.id}"
            
        vote_type = "Like" if self.vote == self.LIKE else "Dislike"
        return f"{vote_type} on {item_str}"

