package client

import (
	"testing"

	"github.com/martketplace-vkr/catalog/domain"
	"github.com/martketplace-vkr/catalog/internal/service/client/dto"
)

func TestFilterCategoriesWithChildrenReturnsOnlyRoots(t *testing.T) {
	rootID := int64(2)
	childID := int64(3)

	categories := domain.CategoryList{
		{ID: rootID, Name: "Электроника"},
		{ID: childID, Name: "Телефон", ParentID: &rootID},
		{ID: 4, Name: "Тапик", ParentID: &childID},
		{ID: 5, Name: "Сенсорный", ParentID: &childID},
	}

	filtered := filterCategories(categories, dto.GetCategoriesRequest{
		IncludeChildren: true,
	})

	if len(filtered) != 1 {
		t.Fatalf("expected one root category, got %d", len(filtered))
	}

	if filtered[0].ID != rootID {
		t.Fatalf("expected root category ID %d, got %d", rootID, filtered[0].ID)
	}

	if len(filtered[0].Children) != 1 {
		t.Fatalf("expected one child category, got %d", len(filtered[0].Children))
	}

	if filtered[0].Children[0].ID != childID {
		t.Fatalf("expected child category ID %d, got %d", childID, filtered[0].Children[0].ID)
	}

	if len(filtered[0].Children[0].Children) != 2 {
		t.Fatalf("expected two nested child categories, got %d", len(filtered[0].Children[0].Children))
	}
}
