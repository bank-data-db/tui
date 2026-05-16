package repo

import "github.com/bank_data_tui/api"

type Cache struct {
	Categories []*api.Category
}

func (s *Cache) EasyCategories(c *api.APIClient) ([]*api.Category, error) {
	if s.Categories != nil {
		return s.Categories, nil
	}

	v, err := c.CategoriesFetch()
	if err != nil {
		return nil, err
	}

	s.Categories = v
	return v, nil
}

func (s *Cache) EasyCatByID(c *api.APIClient, id string) (*api.Category, error) {
	cats, err := s.EasyCategories(c)
	if err != nil {
		return nil, err
	}
	for _, v := range cats {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, nil
}
